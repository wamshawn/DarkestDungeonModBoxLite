export class Error {
    private readonly msg: string

    constructor(msg: string) {
        this.msg = msg
    }

    message(): string {
        return this.msg
    }
}

export class Result<T extends any> {
    private data: T | undefined
    private error?: Error

    static succeed<T>(data: T) {
        const r = new Result<T>();
        r.data = data
        return r;
    }

    static failed<T>(error: Error) {
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

    cause(): string {
        if (this.error) {
            return this.error.message()
        }
        return "no error"
    }

    value(): T {
        if (this.data !== undefined) {
            return this.data
        }
        return {} as T;
    }
}