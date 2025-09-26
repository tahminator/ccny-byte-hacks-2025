import type { ApiResponder } from "@/lib/api/common/responder";
import type { AuthenticationObject } from "@/lib/api/types/auth";

export async function getAuth() {
  const res = await fetch("/api/auth/validate");

  return (await res.json()) as ApiResponder<AuthenticationObject>;
}
