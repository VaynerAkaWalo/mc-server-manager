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
import {ManagerClient, StatsResponse} from "./clients/ManagerClient.ts";
import {cachedCatalog} from "./clients/cachedCatalog.ts";
import {activeServers} from "./servers/activeServers.ts";

const router = createBrowserRouter([
  {
    path: "/",
    element: <App/>,
    loader: () => document.fonts.load("1em Minecraft"),
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
        loader: () => cachedCatalog().catch(() => null)
      },
      {
        path: "/server-templates",
        element: <TemplateList/>,
        loader: () => cachedCatalog().catch(() => null)
      }
    ]
  }
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router}/>
  </StrictMode>,
)
