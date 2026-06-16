---
phase: requirements
title: WC2026 Champion Prediction & Betting
description: Allow users to predict the World Cup 2026 champion team and bet points on it
---

# Requirements & Problem Understanding

## Problem Statement

Hiện tại hệ thống WC2026 chỉ cho phép dự đoán theo từng trận đấu (handicap, tỉ số chính xác). Người dùng muốn có thêm tính năng dự đoán **đội vô địch toàn giải** — đây là loại dự đoán phổ biến và hấp dẫn nhất ở mọi giải đấu lớn. Hiện tại không có cơ chế nào hỗ trợ điều này.

## Goals & Objectives

**Primary goals:**
- Người dùng có thể chọn 1 đội và đặt cược điểm vào đội đó để thắng WC2026
- Admin có thể mở/đóng cửa sổ dự đoán thủ công bất kỳ lúc nào
- Admin set odds cho từng đội (dựa theo sức mạnh), có thể chỉnh sửa sau
- Admin công bố đội vô địch để hệ thống tự động tính điểm và settle

**Secondary goals:**
- Hiển thị bảng odds các đội để người dùng tham khảo trước khi đặt
- Hiển thị ai đang đặt cược đội nào (public leaderboard-style)

**Non-goals:**
- Không hỗ trợ dự đoán nhiều đội cùng lúc (chỉ 1 đội/user)
- Không tự động lấy odds từ API ngoài
- Không tích hợp với match-level prediction/bet hiện tại

## User Stories & Use Cases

**User (người chơi):**
- Là user đã login, tôi muốn xem danh sách các đội với odds tương ứng để quyết định đặt cược
- Là user đã login, tôi muốn chọn 1 đội và cược X điểm (1–5 điểm) vào đội đó
- Là user đã login, tôi muốn xem lại lựa chọn của mình và chỉnh sửa khi cửa sổ còn mở
- Là user đã login, tôi muốn biết kết quả (thắng/thua bao nhiêu điểm) sau khi admin công bố vô địch

**Admin:**
- Là admin, tôi muốn set odds cho từng đội (seed mẫu sẵn, có thể chỉnh)
- Là admin, tôi muốn mở/đóng cửa sổ dự đoán thủ công
- Là admin, tôi muốn nhập đội vô địch và trigger settle tự động cho tất cả user

## Success Criteria

- [ ] User có thể đặt/sửa champion prediction khi cửa sổ mở
- [ ] User nhận điểm đúng công thức: `floor(points × odds)` nếu đúng, mất điểm nếu sai
- [ ] Admin set odds, mở/đóng, và công bố vô địch qua admin panel
- [ ] Sau khi settle, wallet của tất cả user được cập nhật đúng
- [ ] Cửa sổ đóng: user không thể tạo mới hoặc sửa

## Constraints & Assumptions

- Mỗi user chỉ được đặt 1 lần (hoặc sửa nếu cửa sổ còn mở)
- Min cược: 1 điểm — Max cược: 5 điểm (đồng nhất với giới hạn match-level)
- Odds do admin set thủ công; seed mẫu được generate sẵn khi enable feature
- Settle chỉ xảy ra 1 lần — admin không thể undo sau khi đã settle
- Admin phải đóng cửa sổ (`is_open = false`) trước khi settle; nếu không hệ thống báo lỗi
- Điểm **không bị trừ khi đặt cược** — chỉ tính delta (thắng/thua) vào wallet lúc settle
- Tất cả predictions là **public** — mọi user đều thấy ai đặt đội nào và bao nhiêu điểm
- Dùng chung wallet `wc_wallets` với match predictions; settle champion tác động trực tiếp vào balance

## Questions & Open Items

- ~~Khi nào khóa dự đoán?~~ → Admin mở/đóng thủ công
- ~~Odds cố định hay per-team?~~ → Admin set per team, có seed mẫu
- ~~Nếu admin đóng cửa sổ rồi mở lại, user có sửa được không?~~ → **Có**, miễn cửa sổ đang mở
- ~~Settle có tính vào settlement chung (point_rate × balance) không?~~ → **Có**, dùng chung wallet
- ~~Điểm có bị trừ ngay khi đặt cược không?~~ → **Không**, chỉ tính khi admin settle (đồng nhất với match predictions)
- ~~Ai thấy được prediction của người khác?~~ → **Public hoàn toàn** — tất cả user thấy ai đặt đội nào và bao nhiêu điểm
- ~~Admin settle khi cửa sổ vẫn mở thì sao?~~ → **Báo lỗi** — admin phải đóng cửa sổ trước khi settle
