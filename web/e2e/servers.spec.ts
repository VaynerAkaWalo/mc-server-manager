import {expect, Page, test} from "@playwright/test";

const catalog = {
  tiers: [
    {id: "guest", name: "Guest", default: true, maxPlayers: 3, viewDistance: 6, simulationDistance: 4, persistent: false, nameReserved: false, accountRequired: false, available: true},
    {id: "free", name: "Free", default: false, maxPlayers: 5, viewDistance: 8, simulationDistance: 6, persistent: true, nameReserved: true, accountRequired: true, available: true},
  ],
  presets: [],
  options: [{id: "DIFFICULTY", values: ["peaceful", "easy", "normal", "hard"]}],
  versions: {FABRIC: {latestRelease: "1.21.10", items: [{id: "1.21.10", kind: "release"}]}},
}

const server = (id: string, name: string, extra = {}) => ({id, name, hostname: `${id}.mc-server.blamedevs.com`, status: "stopped", tier: "free", saveId: `save-${id}`, saveAvailable: true, saveLocked: false, preset: "", options: {VERSION: "1.21.10", DIFFICULTY: "hard", ONLINE_MODE: "true"}, players: null, createdAt: "2026-11-04T10:00:00Z", expiresAt: "2026-11-05T10:00:00Z", idleShutdownAt: null, ...extra})
const history = [server("running", "Oak valley", {status: "running", players: {online: 2, max: 5}}), server("stopped", "River camp"), server("saving", "Mountain base", {saveLocked: true}), server("expired", "Old cabin", {saveAvailable: false}), server("guest", "Guest visit", {tier: "guest", saveId: undefined, saveAvailable: false})]

type Fixture = {items?: ReturnType<typeof server>[], guest?: boolean, listError?: boolean, serverError?: boolean, saveError?: boolean}

async function mockAPI(page: Page, options: Fixture = {}) {
  let items = options.items ?? []
  let loggedIn = !options.guest
  const serverRequests: Record<string, unknown>[] = []
  const saves: unknown[] = []
  const queries: URLSearchParams[] = []
  await page.route("**/api/v1/**", async route => {
    const request = route.request()
    const url = new URL(request.url())
    const path = url.pathname
    const method = request.method()
    if (path.endsWith("/catalog")) return route.fulfill({json: catalog})
    if (path.endsWith("/sessions/current")) {
      if (method === "DELETE") { loggedIn = false; return route.fulfill({status: 204}) }
      return route.fulfill(loggedIn ? {json: {subject: "alice", email: "alice@example.com", expiresAt: new Date(Date.now() + 3_600_000).toISOString()}} : {status: 401})
    }
    if (path.endsWith("/stats")) return route.fulfill({json: {online: 1, created: 12}})
    if (path === "/api/v1/saves" && method === "POST") {
      saves.push(request.postDataJSON())
      return route.fulfill(options.saveError ? {status: 503} : {status: 201, json: {id: "new-save", name: request.postDataJSON().name, createdAt: "2026-11-05T10:00:00Z"}})
    }
    if (path === "/api/v1/servers" && method === "POST") {
      serverRequests.push(request.postDataJSON())
      if (options.serverError) return route.fulfill({status: 409, json: {detail: "this save is already in use"}})
      const created = server("new-server", request.postDataJSON().name, {...request.postDataJSON(), status: "new"})
      items = [created, ...items]
      return route.fulfill({status: 201, json: created})
    }
    if (path === "/api/v1/servers") {
      queries.push(url.searchParams)
      return route.fulfill(options.listError ? {status: 503} : {json: {items: url.searchParams.get("owner") === "me" ? items : items.filter(item => item.status === "running"), nextCursor: ""}})
    }
    return route.fulfill({json: {items: [], nextCursor: ""}})
  })
  return {serverRequests, saves, queries}
}

async function screenshot(page: Page, name: string) {
  await page.screenshot({path: test.info().outputPath(`${name}.png`), fullPage: true, animations: "disabled"})
  for (const selector of ["main", '[role="region"]']) {
    const region = page.locator(selector)
    if (await region.count() && await region.evaluate(element => element.scrollHeight > element.clientHeight)) {
      const top = await region.evaluate(element => {const old = element.scrollTop; element.scrollTop = element.scrollHeight; return old})
      await page.screenshot({path: test.info().outputPath(`${name}-${selector === "main" ? "page" : "list"}-bottom.png`), fullPage: true, animations: "disabled"})
      await region.evaluate((element, old) => {element.scrollTop = old}, top)
    }
  }
}

test("personal history shows running and previous servers with save availability", async ({page}) => {
  const {queries} = await mockAPI(page, {items: history})
  await page.goto("/my-servers")
  await expect(page.getByRole("heading", {name: "Oak valley"})).toBeVisible()
  await expect(page.getByRole("heading", {name: "River camp"})).toHaveCount(0)
  await screenshot(page, "all-servers")
  await page.getByRole("link", {name: "My servers", exact: true}).click()
  await expect(page.getByRole("heading", {name: "River camp"})).toBeVisible()
  await expect(page.getByText("Saving your world...", {exact: true})).toBeVisible()
  await expect(page.getByText("World expired after 90 days without playing")).toBeVisible()
  await expect(page.getByRole("button", {name: "Start", exact: true})).toHaveCount(3)
  await expect(page.getByRole("button", {name: "Start", exact: true}).nth(0)).toBeEnabled()
  await expect(page.getByRole("button", {name: "Start", exact: true}).nth(1)).toBeDisabled()
  await expect(page.getByRole("button", {name: "Start", exact: true}).nth(2)).toBeDisabled()
  expect(queries.some(query => query.get("owner") === "me" && query.get("status") === "all")).toBe(true)
  await screenshot(page, "my-servers-history")
})

test("restart keeps the save and exact settings and replaces the previous session", async ({page}) => {
  const prior = server("previous", "River camp", {preset: "fabric"})
  const {serverRequests, saves} = await mockAPI(page, {items: [prior, {...prior, id: "older"}]})
  await page.goto("/my-servers?owner=me")
  await expect(page.getByRole("heading", {name: "River camp"})).toHaveCount(1)
  await page.getByRole("button", {name: "Start", exact: true}).press("Enter")
  await expect(page.getByRole("button", {name: "Start", exact: true})).toHaveCount(0)
  await expect(page.getByRole("heading", {name: "Servers", exact: true})).toBeFocused()
  expect(serverRequests).toEqual([{name: prior.name, tier: prior.tier, preset: prior.preset, options: prior.options, saveId: prior.saveId}])
  expect(saves).toEqual([])
  await screenshot(page, "my-servers-restarted")
})

test("restart failure remains inline and restores keyboard focus", async ({page}) => {
  await mockAPI(page, {items: [server("previous", "River camp")], serverError: true})
  await page.goto("/my-servers?owner=me")
  const start = page.getByRole("button", {name: "Start", exact: true})
  await start.press("Enter")
  await expect(page.getByRole("alert")).toContainText("already in use")
  await expect(start).toBeFocused()
  await screenshot(page, "my-servers-start-error")
})

test("history failure can be retried and empty and search states are distinct", async ({page}) => {
  await mockAPI(page, {listError: true})
  await page.goto("/my-servers?owner=me")
  await expect(page.getByRole("alert")).toContainText("Could not load your servers")
  await screenshot(page, "my-servers-load-error")
  await page.route("**/api/v1/servers?**", route => route.fulfill({json: {items: [], nextCursor: ""}}))
  await page.getByRole("button", {name: "Try again"}).click()
  await expect(page.getByText("No servers yet.", {exact: false})).toBeVisible()
  await screenshot(page, "my-servers-empty")
  await page.unroute("**/api/v1/servers?**")
  await mockAPI(page, {items: history})
  await page.reload()
  await page.getByRole("searchbox").fill("missing")
  await expect(page.getByText("No servers match your search.")).toBeVisible()
  await screenshot(page, "my-servers-search-empty")
})

test("logout hides private history immediately", async ({page}) => {
  await mockAPI(page, {items: history})
  await page.goto("/my-servers?owner=me")
  await expect(page.getByRole("heading", {name: "Oak valley"})).toBeVisible()
  await page.getByRole("button", {name: "Account"}).click()
  await page.getByRole("menuitem", {name: "Log out"}).click()
  await expect(page.getByRole("heading", {name: "Oak valley"})).toHaveCount(0)
  await expect(page.getByText("Log in to see your running and previous servers.")).toBeVisible()
  await screenshot(page, "my-servers-login")
})

test("new Free servers automatically create a save and reuse it on retry", async ({page}) => {
  const {serverRequests, saves} = await mockAPI(page, {serverError: true})
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await page.getByLabel("Name", {exact: true}).fill("Oak valley")
  await expect(page.getByText("Your world is saved automatically.", {exact: false})).toBeVisible()
  await screenshot(page, "creator-free")
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect(page.getByRole("alert")).toContainText("already in use")
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect.poll(() => serverRequests.length).toBe(2)
  expect(saves).toEqual([{name: "Oak valley"}])
  expect(serverRequests.every(request => request.saveId === "new-save")).toBe(true)
  await screenshot(page, "creator-error")
})

test("failed automatic save creation prevents server creation", async ({page}) => {
  const {serverRequests} = await mockAPI(page, {saveError: true})
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect(page.getByRole("alert")).toContainText("Failed to create server")
  expect(serverRequests).toEqual([])
  await screenshot(page, "creator-save-error")
})

test("Guest skips saving and logged out Free creation is blocked", async ({page}) => {
  const {serverRequests, saves} = await mockAPI(page, {guest: true})
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeDisabled()
  await screenshot(page, "creator-login")
  await page.getByRole("radio", {name: "Guest", exact: true}).press("Space")
  await screenshot(page, "creator-guest")
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect(page).toHaveURL(/my-servers$/)
  expect(saves).toEqual([])
  expect(serverRequests[0]).not.toHaveProperty("saveId")
})

test("personal history loads all pages before deduplicating saves", async ({page}) => {
  await mockAPI(page)
  const queries: URLSearchParams[] = []
  await page.route("**/api/v1/servers?**", route => {
    const query = new URL(route.request().url()).searchParams
    queries.push(query)
    return route.fulfill({json: {items: query.has("cursor") ? [server("last", "Last page")] : [server("first", "First page")], nextCursor: query.has("cursor") ? "" : "page-two"}})
  })
  await page.goto("/my-servers?owner=me")
  await expect(page.getByRole("heading", {name: "Last page"})).toBeVisible()
  expect(queries.every(query => query.get("owner") === "me" && query.get("status") === "all")).toBe(true)
})

test("mobile history and creator fit the viewport", async ({page}) => {
  await page.setViewportSize({width: 390, height: 844})
  await mockAPI(page, {items: history})
  await page.goto("/my-servers?owner=me")
  await expect(page.getByRole("heading", {name: "River camp"})).toBeVisible()
  await screenshot(page, "my-servers-mobile")
  await page.getByRole("button", {name: "Open navigation"}).click()
  await screenshot(page, "navigation-mobile")
  await page.getByRole("button", {name: "Close menu"}).click()
  await page.getByRole("link", {name: "Create server", exact: true}).click()
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await screenshot(page, "creator-mobile")
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
})

test("home announces returning to saved servers", async ({page}) => {
  await mockAPI(page)
  await page.goto("/")
  await expect(page.getByRole("heading", {name: "Pick up where you left off"})).toBeVisible()
  await screenshot(page, "home")
})

test("account and history loading and account failure remain distinct", async ({page}) => {
  await mockAPI(page)
  let releaseSession: () => void = () => {}
  const sessionReady = new Promise<void>(resolve => {releaseSession = resolve})
  await page.route("**/api/v1/sessions/current", async route => {await sessionReady; return route.fulfill({json: {subject: "alice", expiresAt: new Date(Date.now() + 3_600_000).toISOString()}})})
  let releaseHistory: () => void = () => {}
  const historyReady = new Promise<void>(resolve => {releaseHistory = resolve})
  await page.route("**/api/v1/servers?**", async route => {await historyReady; return route.fulfill({json: {items: [], nextCursor: ""}})})
  await page.goto("/my-servers?owner=me")
  await expect(page.getByText("Checking your account...")).toBeVisible()
  await screenshot(page, "my-servers-account-loading")
  releaseSession()
  await expect(page.getByText("Loading servers...")).toBeVisible()
  await screenshot(page, "my-servers-loading")
  releaseHistory()
  await expect(page.getByText("No servers yet.", {exact: false})).toBeVisible()
  await page.route("**/api/v1/sessions/current", route => route.fulfill({status: 503}))
  await page.reload()
  await expect(page.getByRole("alert")).toContainText("Login status unavailable")
  await screenshot(page, "my-servers-account-error")
})

test("a refresh during restart cannot replace the newly started session", async ({page}) => {
  const prior = server("previous", "River camp")
  await mockAPI(page, {items: [prior]})
  await page.clock.install()
  let releaseStart: () => void = () => {}
  const startReady = new Promise<void>(resolve => {releaseStart = resolve})
  await page.route("**/api/v1/servers", async route => {
    if (route.request().method() !== "POST") return route.fallback()
    await startReady
    return route.fulfill({status: 201, json: server("replacement", prior.name, {...prior, id: "replacement", status: "new"})})
  })
  await page.goto("/my-servers?owner=me")
  await page.getByRole("button", {name: "Start", exact: true}).click()
  await expect(page.getByRole("button", {name: "Starting...", exact: true})).toBeDisabled()
  await screenshot(page, "my-servers-starting")
  let releaseRefresh: () => void = () => {}
  const refreshReady = new Promise<void>(resolve => {releaseRefresh = resolve})
  let refreshing = false
  await page.route("**/api/v1/servers?**", async route => {
    refreshing = true
    await refreshReady
    return route.fulfill({json: {items: [prior], nextCursor: ""}})
  })
  await page.clock.fastForward(10_000)
  await expect.poll(() => refreshing).toBe(true)
  releaseStart()
  await expect(page.getByRole("img", {name: "Starting", exact: true})).toBeVisible()
  const refreshed = page.waitForEvent("requestfinished", request => request.url().includes("/api/v1/servers?"))
  releaseRefresh()
  await refreshed
  await page.clock.runFor(100)
  await expect(page.getByRole("img", {name: "Stopped", exact: true})).toHaveCount(0)
  await expect(page.getByRole("button", {name: "Start", exact: true})).toHaveCount(0)
})

test("unknown server statuses remain readable", async ({page}) => {
  await mockAPI(page, {items: [server("future", "Future session", {status: "hibernating"})]})
  await page.goto("/my-servers?owner=me")
  await expect(page.getByText("hibernating", {exact: true})).toBeVisible()
  await expect(page.getByRole("img", {name: "hibernating"})).toHaveClass(/bg-content-muted/)
})
