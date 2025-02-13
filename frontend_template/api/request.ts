

export async function postJSON<T>(api: string, data: any): Promise<T> {
    const resp = await fetch(api, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
    if (!resp.ok) {
        // status code not 200~299
        throw new Error(`HTTP error! status: ${resp.status}`)
    }
    const json = await resp.json()
    if (json.code !== 0) {
        throw new Error(`API error! code: ${json.code}, message: ${json.message}`)
    }
    return json.result
}
