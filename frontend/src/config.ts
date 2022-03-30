class Expiration {
    id: string = "";
    title: string = "";
    days: number = 0;
}

export class Server {
    id: string = "";
    title: string = "";

    expiration: Array<Expiration> = [];
}

class Features {
    shorten_link: boolean = false;
	notify_mail: boolean = false;
    notify_browser: boolean = false;
	encrypt: boolean = false;
}

export class Config {
    servers: Array<Server> = []
    features: Features
}
