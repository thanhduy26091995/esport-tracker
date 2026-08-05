# Ẩn page World Cup, chỉ để ASEAN Cup

**Date:** 2026-08-05
**Status:** Approved (design)

## Mục tiêu

World Cup 2026 đã hết vai trò; giải đang chạy là ASEAN Cup 2026. Cần làm cho ASEAN Cup
là giải duy nhất mà người dùng thấy được, nhưng **không xoá** hệ thống WC — admin vẫn
phải vào được các trang WC để tra cứu dữ liệu cũ.

## Quyết định phạm vi

| Câu hỏi | Quyết định |
|---------|-----------|
| Mức độ ẩn | Chỉ ẩn khỏi nav menu. Route `/world-cup/*` vẫn hoạt động khi gõ URL trực tiếp. |
| Build nào | Cả 2: site chính (esport tracker) và site `soc` (soc.sitenow.cloud). |
| Widget "WC2026 — Sắp diễn ra" trên Dashboard | Đổi sang fetch/link ASEAN Cup (không xoá widget). |
| Branding site `soc` | Đổi sang ASEAN Cup. |

**Ngoài phạm vi:** router, backend, DB schema, feature flag `/wc/config`. Dữ liệu WC giữ
nguyên toàn bộ. Bật lại WC về sau chỉ cần thêm lại nav item.

## Thiết kế

### 1. Ẩn nav World Cup

`frontend/src/layouts/MainLayout.vue` — bỏ entry `nav.worldCup` khỏi **cả hai** nhánh của
`navigation` (nhánh `isSocSite` và nhánh site chính). Entry `nav.aseanCup` giữ nguyên.

Giữ nguyên (không sửa):

- Route `/world-cup/*` trong `frontend/src/router/index.ts` — admin gõ URL trực tiếp vẫn
  vào được, kể cả `/world-cup/admin`.
- `isWcRoute` vẫn match `/world-cup` → live chat và top-3 honor banner vẫn chạy trên các
  trang WC bị ẩn.
- Key i18n `nav.worldCup` trong `vi.json` / `en.json` — để lại cho lần bật lại.
- CSS `.nav-item--wc` / `.nav-item--wc .nav-icon` trong `MainLayout.vue` — thành dead CSS
  nhưng để lại cùng key i18n cho lần bật lại. `.nav-item-badge` (base) và
  `.nav-item-badge--ac` vẫn đang được badge ASEAN dùng.
- Điều kiện template `item.highlight === 'wc' || item.highlight === 'ac'` — giữ nguyên;
  nhánh `'wc'` chỉ đơn giản là không còn bao giờ khớp.

**Tác dụng phụ đã biết, chấp nhận:** `currentNavItem` dò nav item để render tiêu đề
topbar. Khi ở `/world-cup/*` sẽ không match item nào nên topbar hiện tên app thay vì
"World Cup 2026". Chỉ là cosmetic, trên trang mà giờ chỉ admin mới vào. Không xử lý.

### 2. Widget Dashboard → ASEAN Cup

Backend `/ac/matches` đã tồn tại và public (mirror `/wc` qua
`setupTournamentRoutes("ac", "asean_cup")` trong `backend/internal/api/router.go`), nên
đây là thay đổi thuần frontend.

**`frontend/src/services/wcPublicApi.ts`**

- `listMatchesPublic(filter, prefix)` — thêm tham số `prefix`, dùng helper `publicApiFor(prefix)`
  đã có sẵn trong file. `prefix` là tham số **bắt buộc**, không đặt default: chỉ có một call
  site nên default chỉ tạo chỗ cho sai sót ngầm.
- Xoá `export const wcPublicApi` (axios instance hardcode `/wc`). Đã grep: sau thay đổi
  không còn consumer nào. `WcScheduleView.vue` chỉ import `getStandings` /
  `getTournamentConfig`, cả hai đã nhận `prefix`.

**`frontend/src/views/DashboardView.vue`**

- `fetchUpcomingWcMatches()` gọi `listMatchesPublic({...}, 'ac')`.

**`frontend/src/components/wc/WcUpcomingWidget.vue`**

- `goToSchedule()` → `/asean-cup`.
- 2 router-link → `/asean-cup/predict` và `/asean-cup/login`.
- Title `"⚽ WC2026 — Sắp diễn ra"` → i18n key mới `dashboard.aseanUpcomingTitle`
  ("⚽ ASEAN Cup 2026 — Sắp diễn ra" / "⚽ ASEAN Cup 2026 — Upcoming"). Đưa qua vue-i18n
  vì đang sửa đúng dòng đó (hard rule: mọi UI string mới qua vue-i18n).
- `STAGE_LABELS` hardcode tiếng Việt: **để yên**, không thuộc scope.

Không rename file — `WcUpcomingWidget.vue` giữ tên, đúng convention hiện tại nơi
`WcScheduleView` / `wcAuthStore` / `wcService` đều phục vụ cả hai giải. `wcAuthStore` dùng
chung `wc_token` cho cả hai giải nên `isLoggedIn` / `isAdmin` trong widget vẫn đúng.

Hardcode `'ac'` thay vì thêm prop `tournament: 'wc' | 'ac'`: sau khi ẩn nav WC ở cả hai
build thì không còn consumer nào cần biến thể `wc`. Đổi ngược lại chỉ là 4 dòng.

### 3. Branding site soc

| Chỗ | Trước | Sau |
|-----|-------|-----|
| `vi.json` → `common.appNameSoc` | "Dự Đoán WC" | "Dự Đoán ASEAN Cup" |
| `en.json` → `common.appNameSoc` | "WC Prediction" | "ASEAN Cup Prediction" |
| `vi.json` / `en.json` → `layout.sidebarSubtitleSoc` | "World Cup 2026" | "ASEAN Cup 2026" |
| `frontend/.env.soc` → `VITE_SITE_TITLE` | "WC Prediction 2026" | "ASEAN Cup Prediction 2026" |

### 4. Bổ sung phát sinh khi verify: title trang login

Chạy site `soc` phát hiện card đăng nhập trên `/asean-cup/login` ghi "World Cup 2026" —
`WcLoginView.vue` dùng `t('wc.loginTitle')` hardcode tên giải cho cả hai tournament. Trên
site chỉ còn ASEAN thì đây là landing page duy nhất, nên phải sửa:

- `useTournamentRoutes.ts`: thêm computed `tournamentName` — 'ASEAN Cup 2026' /
  'World Cup 2026', không có emoji (khác `tournamentTitle` sẵn có vì card login đã tự có
  trophy badge riêng).
- `WcLoginView.vue`: `t('wc.loginTitle')` → `tournamentName`.

Không i18n hoá: tên giải là danh từ riêng, `vi.json` và `en.json` vốn đã ghi giống nhau
("World Cup 2026" ở cả hai). Cùng pattern với `tournamentTitle` đang dùng.

Key `wc.loginTitle` / `wc.nav` / `wc.registerTitle` thành key chết — để lại, không xoá.
Tab `wc.analytics.wc2026Tab` đã có `v-if="!isAc"` nên không lộ trên trang ASEAN.

`WcAdminView.vue:36` tự khai lại `tournamentTitle` thay vì dùng composable — duplication
sẵn có, ngoài scope, không sửa.

## Verify

Lưu ý: `npm run type-check` mà `CLAUDE.md` nêu **không tồn tại** trong `package.json`
(chỉ có `dev`, `dev:soc`, `build`, `build:soc`, `preview`). Dùng `npx vue-tsc -b`. Frontend
cũng chưa có test runner nào nên verify là type-check + build + kiểm tay trên browser.

1. `cd frontend && npx vue-tsc -b` — pass.
2. `npm run dev` (site chính): sidebar chỉ còn ASEAN Cup 2026, không còn World Cup;
   widget dashboard hiện match ASEAN Cup và link về `/asean-cup/*`.
3. Build/dev với `.env.soc`: sidebar chỉ còn ASEAN Cup; title + subtitle là ASEAN Cup.
4. Gõ trực tiếp `/world-cup` và `/world-cup/admin` (đăng nhập admin) — vẫn load được, không 404.
5. Widget ẩn khi ASEAN Cup không có match nào trong khung -4h → +72h (điều kiện
   `v-if="upcomingMatches.length > 0"` giữ nguyên).
