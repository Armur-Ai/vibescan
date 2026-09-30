-- INTENTIONALLY VULNERABLE. Do not deploy.

-- V03: no row-level security. Anyone with the anon key can read every row.
create table public.profiles (
  id uuid primary key references auth.users (id),
  email text not null,
  full_name text,
  stripe_customer_id text
);

-- V03: RLS is enabled, but the policy allows everyone to do everything.
create table public.notes (
  id bigint generated always as identity primary key,
  owner_id uuid not null references auth.users (id),
  body text not null
);

alter table public.notes enable row level security;

create policy "notes are accessible" on public.notes
  for all
  to anon, authenticated
  using (true)
  with check (true);
