import {User} from "oidc-client-ts";
import {useEffect, useState} from "react";
import {userManager} from "./userManager.ts";

export function useUser(): User | null {
  const [user, setUser] = useState<User | null>(null)

  useEffect(() => {
    userManager.getUser().then(setUser)

    const clear = () => setUser(null)
    userManager.events.addUserLoaded(setUser)
    userManager.events.addUserUnloaded(clear)

    return () => {
      userManager.events.removeUserLoaded(setUser)
      userManager.events.removeUserUnloaded(clear)
    }
  }, [])

  return user
}
