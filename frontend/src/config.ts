class ExpirationClass {
    tag: string = "";
    title: string = "";
    days: number = 0;
};

export class Config {
    expiration_classes: Array<ExpirationClass> = [];
}
