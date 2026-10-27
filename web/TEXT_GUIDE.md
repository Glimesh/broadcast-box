# where to edit text, colours, music and theme

a little map for the cozy kohi-box frontend. all paths are relative to `web/`.

## text

most of the cozy copy is written directly in the page/component that shows it.
search for the phrase you want to change and edit it in place.

| what you see | file |
| --- | --- |
| nav, search placeholder, ticker, footer, badges, visitor counter | `src/components/rootWrapper/RootWrapper.tsx` |
| home hero, welcome panels, webring / guestbook notes | `src/components/selection/Frontpage.tsx` |
| live channel cards, empty state, loading copy | `src/components/selection/AvailableStreams.tsx` |
| watch page, channel bar, offline / error messages | `src/components/player/PlayerPage.tsx` |
| video HUD, controls, cinema button, offline overlay | `src/components/player/Player.tsx` |
| chat panel, chat placeholder, emote picker | `src/components/player/components/ChatPanel.tsx` |
| stream title / MOTD strip | `src/components/player/components/StreamMOTD.tsx` |
| login / register page | `src/components/account/AuthPage.tsx` |
| account dashboard | `src/components/account/AccountPage.tsx` |
| public profile page | `src/components/account/ProfilePage.tsx` |
| go-live / broadcast page and profile settings | `src/components/broadcast/Broadcast.tsx`, `src/components/broadcast/ProfileSettings.tsx` |
| stream status card | `src/components/broadcast/StreamStatus.tsx` |
| node statistics page | `src/components/statistics/Statistics.tsx` |
| admin login | `src/components/admin/Login.tsx` |
| admin dashboard and menus | `src/components/admin/Frontpage.tsx`, `src/components/admin/menus/*.tsx` |
| shared English labels | `src/locale/en.ts` |

the other locale files (`src/locale/da.ts`, etc.) were intentionally left alone.
if you add brand-new text to a component it will show the same in every language
until it is moved into the locale files.

## theme picker labels

`src/components/shared/ThemeSelector.tsx`

the three choices are `light ♡`, `dark ☾`, and `system`.
the default is light. `system` follows the device / browser preference.

## music player and playlist

`src/config/siteExtras.ts`

- `videoIds` is the youtube playlist. replace / add video ids here.
- `title` and `subtitle` are the little labels in the player.
- `defaultVolume` is 0–100; new visitors start at 30.
- `youtubeUrl` is the “open on youtube” link.

on the first visit the site asks whether it can play music. saying yes unlocks
audio for that click and saves the choice in `bb_music_prompted_v2`; saying no
saves a quiet preference. after that the player tries to start immediately, and
browsers that still block autoplay will start it on the first click/tap instead.
it pauses automatically on watch pages and broadcast pages.
the widget and permission prompt live in `src/components/music/MusicPlayer.tsx`.

## stream titles

channel owners can click `edit ♡` on the stream title strip above the player.
account channels save the title to their normal profile. ad-hoc OBS stream keys
save to `profiles/adhoc/<stream-key>.json` without reserving the key, so the
same stream key keeps working in OBS.

## colours, fonts and decorations

- `src/index.css` — the whole design system, responsive rules, night palette,
  clouds/bubbles, moon/stars, music player styling.
- `index.html` — page title, description, theme colour, favicon, google fonts.

## build / deploy reminder

on the server:

```bash
source /path/to/nvm/nvm.sh
cd /path/to/broadcast-box/web
npm run lint
npm run build
```

the running site serves `web/build`, so the build step is what makes text/design
changes visible on the live site.
