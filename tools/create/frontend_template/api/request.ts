export async function postJSON<T>(api: string, data: any): Promise<T> {
    return request(api, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
}

export async function get<T>(api: string, data?: any): Promise<T> {
    let requestAPI = api
    if (data != null) {
        const params = new URLSearchParams(data)
        requestAPI += `?${params.toString()}`
    }

    return request(requestAPI, {
        method: 'GET',
    })
}

export async function request<T>(api: string, req: RequestInit): Promise<T> {
    const resp = await fetch(api, req)
    if (!resp.ok) {
        // status code not 200~299
        throw new Error(`HTTP error! status: ${resp.status}`)
    }
    const json = await resp.json()
    if (json.code !== 0) {
        throw new Error(`API error! code: ${json.code}, message: ${json.message}`)
    }
    return json.data
}