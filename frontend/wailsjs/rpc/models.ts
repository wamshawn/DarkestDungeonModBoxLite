
export class Failure {
    error: string;
    description: string;

    static createFrom(source: any = {}) {
        return new Failure(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.error = source["error"];
        this.description = source["description"];
    }
}

export class Result<T extends any> {
    private data: T | undefined
    private error?: Array<Failure>

    static succeed<T>(data: T) {
        const r = new Result<T>();
        r.data = data
        return r;
    }

    static failed<T>(error: any) {
        if ('string' === typeof error) error = JSON.parse(error);
        const r = new Result<T>();
        r.error = error
        return r;
    }

    succeed(): boolean {
        return !this.failed();
    }

    failed(): boolean {
        return this.error !== undefined
    }

    cause(): Array<Failure> {
        if (this.error) {
            return this.error
        }
        return new Array<Failure>();
    }

    value(): T {
        if (this.data !== undefined) {
            return this.data
        }
        return {} as T;
    }
}