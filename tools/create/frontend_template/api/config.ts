export async function updateConfig(key: string, config: any) {
    console.log("Updating config:", key, config)
    const response = await fetch('/config/set', {
        method: 'POST',
        body: JSON.stringify({
            key: key,
            value: JSON.stringify(config),
        }),
    });
    return response.json();
}

export async function getConfig<T>(key: string): Promise<T> {
    const response = await fetch('/config/get?' + new URLSearchParams({
        key: key,
    }).toString());
    return await response.json();
}