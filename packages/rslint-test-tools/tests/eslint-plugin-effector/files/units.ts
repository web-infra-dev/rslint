import { createStore } from 'effector';

export const $store = createStore(0);
export const event = $store.updates;
