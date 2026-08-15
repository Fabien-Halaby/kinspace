# KinSpace mobile

React Native app for KinSpace, built with [Expo SDK 54](https://expo.dev) and
[Expo Router](https://docs.expo.dev/router/introduction).

## Features

- **Authentication** — register and login against the KinSpace API; the session token is
  persisted in the platform keychain (`expo-secure-store`) and restored on startup.
- **Home** — current profile and family, with a place for the upcoming feed.
- **Family** — create a family, join one with an invite code, browse the relationship
  graph and add relations between members.

## Structure

```
mobile/
├── app/
│   ├── _layout.tsx         → providers + route guards (auth vs. tabs)
│   ├── (auth)/
│   │   ├── login.tsx
│   │   └── register.tsx
│   └── (tabs)/
│       ├── index.tsx       → Home
│       └── family.tsx      → Family
└── src/
    ├── api/                → typed HTTP client, endpoints, token storage
    ├── auth/               → session context (useAuth)
    └── components/         → reusable themed UI (FormField, PrimaryButton)
```

## Getting started

```bash
npm install
npx expo start
```

The app reaches the API at `http://<expo-host>:8080/api/v1` automatically. To override,
set `EXPO_PUBLIC_API_URL` in `.env.local` (see `.env.example`).

## Checks

```bash
npx tsc --noEmit   # type-check
npx expo lint      # lint
```
