const apiBase = "/api/v1";

export async function apiRequest(req: string, body: object, method = "POST") {
    let resp = await fetch(`${apiBase}/${req}`, {
        method: method,
        mode: "cors",
        headers: {
            "Content-Type": "application/json"
        },
        body: method === "POST" ? JSON.stringify(body) : undefined
    });

    let json = await resp.json();

    if (resp.status !== 200) {
        throw `Failed API request: ${json.error}`;
    }

    return json;
}
