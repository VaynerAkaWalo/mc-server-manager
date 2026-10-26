import {CatalogResponse, ManagerClient} from "./ManagerClient.ts";

let catalog: Promise<CatalogResponse> | undefined

export function cachedCatalog(): Promise<CatalogResponse> {
  catalog ??= ManagerClient.getCatalog()
    .then(({data}) => data)
    .catch(err => {
      catalog = undefined
      throw err
    })

  return catalog
}
