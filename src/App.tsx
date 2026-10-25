import './App.css'
import {Footer} from "./base/Footer.tsx";
import {Outlet, useOutletContext} from "react-router-dom";
import {Notification} from "./base/Notification.tsx";
import {useEffect, useState} from "react";
import {Header} from "./base/Header.tsx";
import {createTheme, ThemeProvider} from "@mui/material";

type NotificationContext =  { setNotificationMessage: React.Dispatch<React.SetStateAction<string>> }

function App() {
  const [notificationMessage, setNotificationMessage] = useState<string>("")
  const [showNotification, setShowNotification] = useState<boolean>(false)

  const theme = createTheme({
    typography: {
      fontFamily: "Minecraft, serif",
    }
  })

  useEffect(() => {
    if (notificationMessage === "") {
      return
    }
    const current = notificationMessage;
    setShowNotification(true)
    setTimeout(() => {
      if (notificationMessage == current) {
        setShowNotification(false)
        setNotificationMessage("")
      }
    }, 3500)
  }, [notificationMessage])

  return (
    <>
      <ThemeProvider theme={theme}>
        <div className="h-screen flex flex-col text-gray-700">
          <Header/>
          <div className="flex h-full justify-center items-center">
            <Outlet context={{ setNotificationMessage } satisfies NotificationContext } />
          </div>
          {showNotification && <Notification message={notificationMessage}/>}
        </div>
        <Footer/>
      </ThemeProvider>
    </>
  )
}

export function UpdateNotification() {
  return useOutletContext<NotificationContext>()
}

export default App
