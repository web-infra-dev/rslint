/** Canonical identity while a plugin can still observe the object. Dead values
 * and their keys are released without retaining a generation's visited graph. */
export class WeakObjectCache<K, V extends object> {
  private readonly values = new Map<K, WeakRef<V>>();
  private readonly collected = new FinalizationRegistry<K>((key) => {
    if (!this.values.get(key)?.deref()) this.values.delete(key);
  });
  get(key: K): V | undefined {
    const reference = this.values.get(key),
      value = reference?.deref();
    if (reference && !value) this.values.delete(key);
    return value;
  }
  set(key: K, value: V): void {
    this.values.set(key, new WeakRef(value));
    this.collected.register(value, key);
  }
  clear(): void {
    this.values.clear();
  }
}
