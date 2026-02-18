import "@testing-library/jest-dom/vitest";

if (typeof window !== "undefined" && typeof window.localStorage?.setItem !== "function") {
  const storage = new Map<string, string>();
  Object.defineProperty(window, "localStorage", {
    configurable: true,
    value: {
      clear() {
        storage.clear();
      },
      getItem(key: string) {
        return storage.get(key) ?? null;
      },
      key(index: number) {
        return Array.from(storage.keys())[index] ?? null;
      },
      removeItem(key: string) {
        storage.delete(key);
      },
      setItem(key: string, value: string) {
        storage.set(String(key), String(value));
      },
      get length() {
        return storage.size;
      },
    },
  });
}
