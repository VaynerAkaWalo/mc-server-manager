import {expect, Page, test} from "@playwright/test";

const world = {id: "5b801cf8-e686-4a82-a670-9716b1862795", name: "Oak valley", createdAt: "2026-11-04T10:00:00Z"}
const catalog = {
  tiers: [
    {id: "guest", name: "Guest", default: true, maxPlayers: 3, viewDistance: 6, simulationDistance: 4, persistent: false, nameReserved: false, accountRequired: false, available: true},
    {id: "free", name: "Free", default: false, maxPlayers: 5, viewDistance: 8, simulationDistance: 6, persistent: true, nameReserved: true, accountRequired: true, available: true},
  ],
  presets: [],
  options: [{id: "DIFFICULTY", values: ["peaceful", "easy", "normal", "hard"]}],
  versions: {FABRIC: {latestRelease: "1.21.10", items: [{id: "1.21.10", kind: "release"}]}},
}

type Fixture = {
  worlds?: typeof world[],
  guest?: boolean,
  listError?: boolean,
  mutationError?: boolean,
  serverError?: boolean,
}

async function mockAPI(page: Page, options: Fixture = {}) {
  let worlds = options.worlds ?? []
  let loggedIn = !options.guest
  const serverRequests: unknown[] = []
  const mutations: string[] = []

  await page.route("**/api/v1/**", async route => {
    const request = route.request()
    const path = new URL(request.url()).pathname
    const method = request.method()

    if (path.endsWith("/catalog")) return route.fulfill({json: catalog})
    if (path.endsWith("/sessions/current")) {
      if (method === "DELETE") {
        loggedIn = false
        return route.fulfill({status: 204})
      }
      return route.fulfill(loggedIn ? {json: {subject: "alice", email: "alice@example.com", expiresAt: new Date(Date.now() + 3_600_000).toISOString()}} : {status: 401})
    }
    if (path.endsWith("/stats")) return route.fulfill({json: {online: 0, created: 12}})
    if (path === "/api/v1/saves" && method === "GET") {
      return options.listError ? route.fulfill({status: 503, json: {detail: "Storage unavailable"}}) : route.fulfill({json: {items: worlds, nextCursor: ""}})
    }
    if (path.startsWith("/api/v1/saves") && method !== "GET") {
      mutations.push(method)
      if (options.mutationError) return route.fulfill(method === "DELETE"
        ? {status: 409, json: {detail: "a server is still using this save, try again in a few minutes"}}
        : {status: 503, json: {detail: "Storage unavailable"}})
      if (method === "DELETE") {
        worlds = worlds.filter(item => !path.endsWith(item.id))
        return route.fulfill({status: 204})
      }
      const updated = {...world, id: method === "POST" ? crypto.randomUUID() : path.split("/").at(-1)!, name: request.postDataJSON().name}
      worlds = method === "POST" ? [...worlds, updated] : worlds.map(item => item.id === updated.id ? updated : item)
      return route.fulfill({status: method === "POST" ? 201 : 200, json: updated})
    }
    if (path === "/api/v1/servers" && method === "POST") {
      serverRequests.push(request.postDataJSON())
      return options.serverError ? route.fulfill({status: 409, json: {detail: "this save is already in use"}}) : route.fulfill({status: 201, json: {}})
    }
    return route.fulfill({json: {items: [], nextCursor: ""}})
  })

  return {serverRequests, mutations}
}

async function screenshot(page: Page, name: string) {
  await page.screenshot({path: test.info().outputPath(`${name}.png`), fullPage: true, animations: "disabled"})
  const main = page.locator("main")
  const scrollable = await main.evaluate(element => element.scrollHeight > element.clientHeight)
  if (scrollable) {
    const scrollTop = await main.evaluate(element => {const top = element.scrollTop; element.scrollTop = element.scrollHeight; return top})
    await page.screenshot({path: test.info().outputPath(`${name}-bottom.png`), fullPage: true, animations: "disabled"})
    await main.evaluate((element, top) => { element.scrollTop = top }, scrollTop)
  }
}

test("create, rename and delete a world only after confirmation", async ({page}) => {
  const {mutations} = await mockAPI(page)
  await page.goto("/my-worlds")
  await expect(page.getByText("No worlds yet.", {exact: false})).toBeVisible()
  await screenshot(page, "worlds-empty")

  await page.getByLabel("New world name").fill("Oak valley")
  await page.getByRole("button", {name: "Create world", exact: true}).click()
  await expect(page.getByRole("heading", {name: "Oak valley"})).toBeVisible()
  await screenshot(page, "worlds-list")

  await page.getByRole("button", {name: "Rename", exact: true}).click()
  await expect(page.getByLabel("World name", {exact: true})).toBeFocused()
  await page.getByLabel("World name", {exact: true}).fill("Birch valley")
  await screenshot(page, "worlds-rename")
  await page.getByRole("button", {name: "Save name"}).click()
  await expect(page.getByRole("heading", {name: "Birch valley"})).toBeVisible()

  await page.getByRole("button", {name: "Delete", exact: true}).click()
  await expect(page.getByText("This cannot be undone.", {exact: false})).toBeVisible()
  await expect(page.getByRole("button", {name: "Cancel"})).toBeFocused()
  expect(mutations).toEqual(["POST", "PATCH"])
  await screenshot(page, "worlds-delete-confirmation")
  await page.getByRole("button", {name: "Delete world", exact: true}).click()
  await expect(page.getByText("No worlds yet.", {exact: false})).toBeVisible()
  expect(mutations).toEqual(["POST", "PATCH", "DELETE"])
})

test("an account with three worlds can create another", async ({page}) => {
  await mockAPI(page, {worlds: [world, {...world, id: "second", name: "River camp"}, {...world, id: "third", name: "Mountain base"}]})
  await page.goto("/my-worlds")
  await expect(page.getByLabel("New world name")).toBeEnabled()
  await page.getByLabel("New world name").fill("Fourth world")
  await expect(page.getByRole("button", {name: "Create world", exact: true})).toBeEnabled()
  await page.getByRole("button", {name: "Create world", exact: true}).click()
  await expect(page.getByText("4 worlds", {exact: true})).toBeVisible()
  await screenshot(page, "worlds-no-cap")
})

test("a world still in use remains listed after deletion fails", async ({page}) => {
  await mockAPI(page, {worlds: [world], mutationError: true})
  await page.goto("/my-worlds")
  await page.getByRole("button", {name: "Delete", exact: true}).click()
  await page.getByRole("button", {name: "Delete world", exact: true}).click()
  await expect(page.getByRole("alert")).toContainText("still using this save")
  await expect(page.getByRole("heading", {name: world.name})).toBeVisible()
  await screenshot(page, "worlds-in-use")
})

test("failed world loading is distinct from an empty account and can be retried", async ({page}) => {
  await mockAPI(page, {listError: true})
  await page.goto("/my-worlds")
  await expect(page.getByRole("alert")).toContainText("Could not load worlds")
  await expect(page.getByRole("button", {name: "Create world", exact: true})).toHaveCount(0)
  await screenshot(page, "worlds-error")
  await page.route("**/api/v1/saves", route => route.fulfill({json: {items: [world], nextCursor: ""}}))
  await page.getByRole("button", {name: "Try again"}).click()
  await expect(page.getByRole("heading", {name: world.name})).toBeVisible()
})

test("starting a saved world selects Free and sends its save id", async ({page}) => {
  const {serverRequests} = await mockAPI(page, {worlds: [world]})
  await page.goto("/my-worlds")
  await page.getByRole("link", {name: "Start server"}).click()
  await expect(page.getByRole("radio", {name: "Free", exact: true})).toBeChecked()
  await expect(page.getByText(world.name, {exact: true})).toBeVisible()
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeEnabled()
  await screenshot(page, "creator-world-selected")
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect(page).toHaveURL(/my-servers$/)
  expect(serverRequests).toEqual([expect.objectContaining({tier: "free", saveId: world.id})])
})

test("Free requires a world while Guest never sends a save id", async ({page}) => {
  const {serverRequests} = await mockAPI(page)
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await expect(page.getByText("Create a world before", {exact: false})).toBeVisible()
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeDisabled()
  await screenshot(page, "creator-no-worlds")
  await page.getByRole("radio", {name: "Guest", exact: true}).press("Space")
  await screenshot(page, "creator-guest")
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect(page).toHaveURL(/my-servers$/)
  expect(serverRequests).toEqual([expect.objectContaining({tier: "guest"})])
  expect(serverRequests[0]).not.toHaveProperty("saveId")
})

test("world picker works with the keyboard and surfaces a save conflict", async ({page}) => {
  await mockAPI(page, {worlds: [world], serverError: true})
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await page.getByLabel("World", {exact: true}).focus()
  await page.keyboard.press("ArrowDown")
  await page.keyboard.press("Enter")
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeEnabled()
  await page.getByRole("button", {name: "Create", exact: true}).click()
  await expect(page.getByRole("alert")).toContainText("already in use")
  await screenshot(page, "creator-in-use")
})

test("world picker handles loading failure without enabling creation", async ({page}) => {
  await mockAPI(page, {listError: true})
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await expect(page.getByRole("alert")).toContainText("Could not load worlds")
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeDisabled()
  await screenshot(page, "creator-worlds-error")
})

test("guests see a login prompt and logging out removes saved worlds immediately", async ({page}) => {
  await mockAPI(page, {worlds: [world]})
  await page.goto("/my-worlds")
  await expect(page.getByRole("heading", {name: world.name})).toBeVisible()
  await page.getByRole("button", {name: "Account"}).click()
  await page.getByRole("menuitem", {name: "Log out"}).click()
  await expect(page.getByRole("heading", {name: world.name})).toHaveCount(0)
  await expect(page.getByText("Log in to keep your worlds")).toBeVisible()
  await expect(page.getByRole("menu")).toHaveCount(0)
  await screenshot(page, "worlds-logged-out")
})

test("logged-out Free creation is blocked", async ({page}) => {
  await mockAPI(page, {guest: true})
  await page.goto("/server-creator")
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await expect(page.getByText("to keep your world on a Free server.", {exact: false})).toBeVisible()
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeDisabled()
  await screenshot(page, "creator-login")
})

test("mobile worlds, navigation and creator fit the viewport", async ({page}) => {
  await page.setViewportSize({width: 390, height: 844})
  await mockAPI(page, {worlds: [world]})
  await page.goto("/my-worlds")
  await expect(page.getByRole("heading", {name: world.name})).toBeVisible()
  await screenshot(page, "worlds-mobile")
  await page.getByRole("button", {name: "Open navigation"}).click()
  await expect(page.getByRole("link", {name: "My worlds"})).toBeVisible()
  await screenshot(page, "worlds-mobile-navigation")
  await page.getByRole("button", {name: "Close menu"}).click()
  await page.getByRole("link", {name: "Start server"}).click()
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeEnabled()
  await screenshot(page, "creator-mobile")
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
})

test("home announces persistent worlds", async ({page}) => {
  await mockAPI(page)
  await page.goto("/")
  await expect(page.getByRole("heading", {name: "Your worlds"})).toBeVisible()
  await screenshot(page, "home")
})

test("world and account loading never show an empty account", async ({page}) => {
  await mockAPI(page)
  let releaseSession: () => void = () => {}
  const sessionReady = new Promise<void>(resolve => { releaseSession = resolve })
  await page.route("**/api/v1/sessions/current", async route => {
    await sessionReady
    return route.fulfill({json: {subject: "alice", expiresAt: new Date(Date.now() + 3_600_000).toISOString()}})
  })
  let releaseWorlds: () => void = () => {}
  const worldsReady = new Promise<void>(resolve => { releaseWorlds = resolve })
  await page.route("**/api/v1/saves", async route => {
    await worldsReady
    return route.fulfill({json: {items: [], nextCursor: ""}})
  })
  await page.goto("/my-worlds")
  await expect(page.getByText("Checking your account...")).toBeVisible()
  await screenshot(page, "worlds-account-loading")
  releaseSession()
  await expect(page.getByText("Loading worlds...")).toBeVisible()
  await screenshot(page, "worlds-loading")
  releaseWorlds()
  await expect(page.getByText("No worlds yet.", {exact: false})).toBeVisible()
})

test("failed create and rename preserve the entered world name", async ({page}) => {
  await mockAPI(page, {worlds: [world], mutationError: true})
  await page.goto("/my-worlds")
  await page.getByLabel("New world name").fill("Another valley")
  await page.getByRole("button", {name: "Create world", exact: true}).click()
  await expect(page.getByRole("alert")).toBeVisible()
  await expect(page.getByLabel("New world name")).toHaveValue("Another valley")
  await screenshot(page, "worlds-create-error")
  await page.getByRole("button", {name: "Rename", exact: true}).click()
  await page.getByLabel("World name", {exact: true}).fill("New name")
  await page.getByRole("button", {name: "Save name"}).click()
  await expect(page.getByLabel("World name", {exact: true})).toHaveValue("New name")
  await expect(page.getByRole("heading", {name: world.name})).toBeVisible()
  await screenshot(page, "worlds-rename-error")
})

test("unavailable login status keeps account worlds hidden", async ({page}) => {
  await mockAPI(page, {worlds: [world]})
  await page.route("**/api/v1/sessions/current", route => route.fulfill({status: 503}))
  await page.goto("/my-worlds")
  await expect(page.getByRole("alert")).toContainText("Login status unavailable")
  await expect(page.getByRole("button", {name: "Create world", exact: true})).toHaveCount(0)
  await screenshot(page, "worlds-account-error")
})

test("world actions restore keyboard focus after save, cancel and delete", async ({page}) => {
  await mockAPI(page, {worlds: [world]})
  await page.goto("/my-worlds")
  const rename = page.getByRole("button", {name: "Rename", exact: true})
  const remove = page.getByRole("button", {name: "Delete", exact: true})
  await rename.click()
  await page.getByRole("button", {name: "Cancel"}).click()
  await expect(rename).toBeFocused()
  await remove.click()
  await page.getByRole("button", {name: "Cancel"}).click()
  await expect(remove).toBeFocused()
  await rename.click()
  await page.getByRole("button", {name: "Save name"}).click()
  await expect(rename).toBeFocused()
  await remove.click()
  await page.getByRole("button", {name: "Delete world", exact: true}).click()
  await expect(page.getByRole("heading", {name: "My worlds", exact: true})).toBeFocused()
  await screenshot(page, "worlds-deleted-focus")
})

test("a transient session refresh failure preserves a known account's rename form", async ({page}) => {
  await mockAPI(page, {worlds: [world]})
  await page.goto("/my-worlds")
  await page.getByRole("button", {name: "Rename", exact: true}).click()
  await page.getByLabel("World name", {exact: true}).fill("Unsaved name")
  await page.route("**/api/v1/sessions/current", route => route.fulfill({status: 503}))
  await page.evaluate(() => window.dispatchEvent(new Event("focus")))
  await expect(page.getByRole("status")).toContainText("Login status unavailable")
  await expect(page.getByLabel("World name", {exact: true})).toHaveValue("Unsaved name")
  await screenshot(page, "worlds-session-refresh-error")
  await page.getByRole("link", {name: "Creator", exact: true}).click()
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await page.getByLabel("World", {exact: true}).focus()
  await page.keyboard.press("ArrowDown")
  await page.keyboard.press("Enter")
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeEnabled()
  await screenshot(page, "creator-session-refresh-error")
})


test("all pages of worlds remain available after the count cap is removed", async ({page}) => {
  await mockAPI(page)
  const worlds = Array.from({length: 21}, (_, i) => ({...world, id: `world-${i}`, name: `World ${i + 1}`}))
  await page.route("**/api/v1/saves**", route => {
    const cursor = new URL(route.request().url()).searchParams.get("cursor")
    return route.fulfill({json: {items: cursor ? worlds.slice(20) : worlds.slice(0, 20), nextCursor: cursor ? "" : "next-page"}})
  })
  await page.goto("/my-worlds")
  await expect(page.getByRole("heading", {name: "World 21", exact: true})).toBeVisible()
  await page.getByRole("link", {name: "Creator", exact: true}).click()
  await page.getByRole("radio", {name: "Free", exact: true}).press("Space")
  await page.getByLabel("World", {exact: true}).fill("World 21")
  await page.keyboard.press("ArrowDown")
  await page.keyboard.press("Enter")
  await expect(page.getByRole("button", {name: "Create", exact: true})).toBeEnabled()
})
