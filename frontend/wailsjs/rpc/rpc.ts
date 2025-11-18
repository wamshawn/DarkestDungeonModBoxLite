/// <reference lib="es2015.promise" />

import {Result} from "./models";

export type NoParamHandler<R extends any> = () => Promise<R>;

export type Handler<P extends any, R extends any> = (param: P) => Promise<R>;

export const rpc = async <P extends any, R extends any>(h: Handler<P, R> | NoParamHandler<R>, param?: P): Promise<Result<R>> => {
    try {
        if (param) {
            const v = await h(param)
            return Result.succeed(v);
        } else {
            const nph = h as NoParamHandler<R>
            const v = await nph()
            return Result.succeed(v);
        }
    } catch (e: any) {
        return Result.failed(e)
    }
}
