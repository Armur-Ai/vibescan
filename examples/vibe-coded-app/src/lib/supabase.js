// INTENTIONALLY VULNERABLE. Do not deploy. The key below is a fake with an invalid signature.
"use client";

import { createClient } from "@supabase/supabase-js";

// V01: a service-role key (bypasses row-level security) exposed to the browser.
// The NEXT_PUBLIC_ prefix inlines it into the client bundle, and the fallback
// hardcodes it in source.
const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL;
const serviceRoleKey =
  process.env.NEXT_PUBLIC_SUPABASE_SERVICE_ROLE_KEY ||
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImV4YW1wbGVwcm9qZWN0Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTcwMDAwMDAwMCwiZXhwIjoyMDAwMDAwMDAwfQ.FAKE-SIGNATURE-FOR-VIBESCAN-FIXTURE";

export const supabase = createClient(supabaseUrl, serviceRoleKey);
