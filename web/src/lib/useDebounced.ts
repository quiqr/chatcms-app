import { useEffect, useState } from "react";

export function useDebounced<T>(value: T, wait: number): T {
  const [settled, setSettled] = useState(value);
  const said = JSON.stringify(value);

  useEffect(() => {
    const timer = setTimeout(() => setSettled(JSON.parse(said) as T), wait);
    return () => clearTimeout(timer);
  }, [said, wait]);

  return settled;
}
