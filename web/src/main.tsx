import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import {createBrowserRouter, RouterProvider} from "react-router-dom";
import {ServerList} from "./servers/ServerList.tsx";
import {ServerCreator} from "./creator/ServerCreator.tsx";
import {TemplateList} from "./templates/TemplateList.tsx";
import {Home} from "./home/Home.tsx";
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
        loader: async (): Promise<StatsResponse> => {
          try {
            const {data} = await ManagerClient.getStats()
            return data
          } catch {
            return {online: 0, created: 0}
          }
        }
      },
      {
        path: "/my-servers",
        element: <ServerList/>,
        loader: () => activeServers().catch(() => null)
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
