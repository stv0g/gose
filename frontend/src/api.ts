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

    if (resp.status !== 200) {
        throw resp;
    }

    return resp.json();
}
