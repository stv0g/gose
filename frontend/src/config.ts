// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

class Expiration {
    id: string = "";
    title: string = "";
    days: number = 0;
}

export class Server {
    id: string = "";
    title: string = "";

    part_size: number = 0;
    max_upload_size: number = 0;

    expiration: Array<Expiration> = [];
}

class Features {
    short_url: boolean = false;
	notify_mail: boolean = false;
    notify_browser: boolean = false;
}

class Build {
    version: string
    commit: string
    state: string
}

export class Config {
    servers: Array<Server> = []
    features: Features
    build: Build
}
