// GENERATED FILE, DO NOT EDIT

export const post = function<T,U>(path: string, req?: T): Promise<U> {
  return new Promise<U>((resolve, reject) =>  {
    const { host, protocol } = window.location
    const url = `${protocol}//${host}${path}`

    let fetchArg = {
      method: 'POST',
      credentials: 'same-origin',
    } as RequestInit

    if (req) {
      fetchArg.headers = { 'Content-Type': 'application/json' }
      fetchArg.body = JSON.stringify(req)
    }

    fetch(url, fetchArg).then(resp => resp.json().then(obj => resolve(obj as U)))
  })
}


export interface reqType {
  A: number
  B: string
}

export interface respType {
  X: string
  Y: number
}


export const Server = {
  foo_bar: (req: reqType) => {
    return post('/s/foo_bar', req) as Promise<respType>
  },
}
