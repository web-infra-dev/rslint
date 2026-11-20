// A reduced copy of effector's public unit types. Rules match units by type
// name plus the `effector` package origin, which `declare module` provides
// without a node_modules tree.
declare module 'effector' {
  export type Subscription = {
    (): void;
    unsubscribe(): void;
  };

  export interface Unit<T> {
    readonly kind: string;
    readonly __: T;
  }

  export interface Event<Payload> extends Unit<Payload> {
    watch(watcher: (payload: Payload) => any): Subscription;
    map<T>(fn: (payload: Payload) => T): Event<T>;
  }

  export interface EventCallable<Payload> extends Event<Payload> {
    (payload: Payload): Payload;
  }

  export interface Store<State> extends Unit<State> {
    getState(): State;
    watch(watcher: (state: State) => any): Subscription;
    updates: Event<State>;
  }

  export interface StoreWritable<State> extends Store<State> {}

  export interface Effect<Params, Done, Fail = Error> extends Unit<Params> {
    (payload: Params): Promise<Done>;
    watch(watcher: (payload: Params) => any): Subscription;
    done: Event<{ params: Params; result: Done }>;
    doneData: Event<Done>;
    fail: Event<{ params: Params; error: Fail }>;
    failData: Event<Fail>;
    finally: Event<{ params: Params; status: 'done' | 'fail' }>;
    pending: Store<boolean>;
  }

  export interface Domain extends Unit<any> {
    watch(watcher: (payload: any) => any): Subscription;
    createEvent<E = void>(): EventCallable<E>;
  }

  export function createEvent<E = void>(): EventCallable<E>;
  export function createStore<State>(defaultState: State): StoreWritable<State>;
  export function createEffect<Params = void, Done = void, Fail = Error>(
    handler?: (params: Params) => Done | Promise<Done>,
  ): Effect<Params, Done, Fail>;
  export function createDomain(): Domain;
  export function sample<T, R = T>(config: {
    clock: Unit<T>;
    fn?: (value: T) => R;
    filter?: (value: T) => boolean;
    target?: Unit<R>;
  }): Event<R>;
  export function guard<T>(config: {
    clock: Unit<T>;
    filter: (value: T) => boolean;
  }): Event<T>;
  export function createWatch<T>(config: {
    unit: Unit<T>;
    fn: (value: T) => any;
  }): Subscription;
}
