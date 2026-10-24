import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import {createBrowserRouter, RouterProvider} from "react-router-dom";
import {ServerList} from "./servers/ServerList.tsx";
import {ServerCreator} from "./creator/ServerCreator.tsx";
import {TemplateList} from "./templates/TemplateList.tsx";

const router = createBrowserRouter([
  {
    path: "/",
    element: <App/>,
    children: [
      {
        path: "/",
        element: <ServerList/>
      },
      {
        path: "/server-creator",
        element: <ServerCreator/>
      },
      {
        path: "/server-templates",
        element: <TemplateList/>
      }
    ]
  }
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router}/>
  </StrictMode>,
)
