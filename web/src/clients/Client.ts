import axios from 'axios';

const GET_TIMEOUT_MS = 15_000

class Client {

  public get = async (url: string) => {
    return axios.get(url, {timeout: GET_TIMEOUT_MS})
  }

  public post = async <T=never>(url: string, request: T): Promise<never> => {
    return axios.post(url, request)
  }
}

export const HTTPClient = new Client()
