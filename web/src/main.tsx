import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import {createBrowserRouter, redirect, RouterProvider} from "react-router-dom";
import {ServerList} from "./servers/ServerList.tsx";
import {ServerAdmin} from "./servers/ServerAdmin.tsx";
import {ServerCreator} from "./creator/ServerCreator.tsx";
import {TemplateList} from "./templates/TemplateList.tsx";
import {Home, HomeData} from "./home/Home.tsx";
import {Spinner} from "./base/Spinner.tsx";
import {CatalogResponse, ManagerClient, StatsResponse} from "./clients/ManagerClient.ts";
import {activeServers} from "./servers/activeServers.ts";

async function loadCatalog(): Promise<CatalogResponse | null> {
  try {
    const {data} = await ManagerClient.getCatalog()
    return data
  } catch {
    return null
  }
}

async function loadStats(): Promise<StatsResponse> {
  try {
    const {data} = await ManagerClient.getStats()
    return data
  } catch {
    return {online: 0, created: 0}
  }
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <App/>,
    loader: () => document.fonts.load("1em Minecraft").catch(() => []),
    hydrateFallbackElement: <div className="h-screen flex items-center justify-center"><Spinner/></div>,
    children: [
      {
        path: "/",
        element: <Home/>,
        loader: async (): Promise<HomeData> => {
          const [stats, servers] = await Promise.all([
            loadStats(),
            activeServers().catch(() => null),
          ])
          return {stats, servers}
        }
      },
      {
        path: "/my-servers",
        element: <ServerList/>,
        loader: ({request}) => new URL(request.url).searchParams.get("owner") === "me" ? redirect("/manage-servers") : null,
      },
      {
        path: "/manage-servers",
        element: <ServerList managed/>,
        loader: loadCatalog
      },
      {
        path: "/manage-servers/:serverId",
        element: <ServerAdmin/>,
        loader: loadCatalog
      },
      {
        path: "/server-creator",
        element: <ServerCreator/>,
        loader: loadCatalog
      },
      {
        path: "/server-templates",
        element: <TemplateList/>,
        loader: loadCatalog
      }
    ]
  }
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router}/>
  </StrictMode>,
)
