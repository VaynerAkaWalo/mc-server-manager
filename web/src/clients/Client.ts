import axios from 'axios';

class Client {

  public get = async (url: string) => {
    return axios.get(url)
  }

  public post = async <T=never>(url: string, request: T): Promise<never> => {
    return axios.post(url, request)
  }
}

export const HTTPClient = new Client()
