# World Cup Prediction Analytics System

## 1. Tổng quan hệ thống

Hệ thống analytics được chia thành 3 lớp chính:

- **My Analytics (Cá nhân)**
- **Community Analytics (Toàn bộ user)**
- **Compare Layer (You vs Community)**

Mục tiêu:
- Tăng mức độ quay lại của user
- Tạo insight cá nhân hóa
- Hiển thị xu hướng dự đoán theo thời gian thực
- Gamification hành vi dự đoán

---

# 2. My Analytics (Cá nhân)

## 2.1 Prediction Profile (Phong cách dự đoán)

Phân loại người dùng theo hành vi:

- Aggressive Predictor (thích cửa dưới)
- Conservative Predictor (an toàn)
- Draw Master (hay chọn hòa)
- Goal Hunter (thích nhiều bàn thắng)
- Favorite Hunter (thích đội mạnh)
- Underdog Lover (ngược kèo)

Output:
- Label + mô tả hành vi

---

## 2.2 Accuracy Timeline (Độ chính xác theo thời gian)

Biểu đồ line chart:

- Accuracy theo ngày / tuần / vòng đấu
- Filter:
  - 7 ngày
  - 30 ngày
  - Tournament

---

## 2.3 Prediction Heatmap

- Theo ngày trong tuần
- Theo giờ trong ngày

Mục tiêu:
- Xác định thói quen user activity

---

## 2.4 Favorite Teams

Thống kê đội được chọn nhiều nhất:

- Brazil
- France
- Spain
- Argentina

---

## 2.5 Favorite Score

Các tỷ số user hay chọn:

- 2-1
- 1-0
- 2-0
- 3-1

---

## 2.6 Average Goals Prediction

So sánh:

- User average predicted goals
- Community average

Insight:
- User thiên về trận nhiều hay ít bàn thắng

---

## 2.7 Home/Away Bias

Phân tích thiên kiến:

- Home Win %
- Away Win %
- Draw %

---

## 2.8 Bet Type Preference

- Win
- Draw
- Lose

---

## 2.9 Upset Hunter Score

Tỷ lệ chọn đội cửa dưới:

- User vs Community
- Ranking percentile

---

## 2.10 Streak System

- Win streak
- Lose streak
- Consecutive correct predictions

---

## 2.11 Prediction Personality (AI Insight)

AI generate:

- Hành vi dự đoán
- Xu hướng tâm lý
- Thói quen chọn kèo

---

## 2.12 Confidence vs Accuracy

Nếu có confidence input:

- Confidence trung bình
- Accuracy tương ứng
- Overconfidence detection

---

# 3. Community Analytics

## 3.1 Trending Teams

- Đội tăng giảm độ phổ biến theo thời gian
- Bullish / Bearish trend

---

## 3.2 Trending Scorelines

- Tỷ số phổ biến nhất toàn cộng đồng

---

## 3.3 Sentiment Analysis

- Bullish / Bearish theo đội
- % cộng đồng tin thắng

---

## 3.4 Community Mood

- Aggressive / Balanced / Defensive

---

## 3.5 Prediction Distribution

- Home / Draw / Away distribution

---

## 3.6 Average Goals Trend

- Xu hướng bàn thắng trung bình theo thời gian

---

## 3.7 Prediction Time Heatmap

- Giờ nào user thường predict nhiều nhất

---

## 3.8 Weekly Evolution

- Accuracy / behavior theo tuần

---

## 3.9 Top Predictors

Ranking theo:

- Accuracy
- Activity
- Upset picks
- Score precision

---

## 3.10 Community Personality

Insight tổng thể:

- Community đang thiên tấn công hay phòng thủ
- Xu hướng thay đổi theo giải đấu

---

# 4. Compare Layer (You vs Community)

## 4.1 Key Comparison Table

So sánh:

- Home win %
- Away win %
- Draw %
- Average goals
- Accuracy
- Underdog rate

---

## 4.2 Behavioral Difference

- Bạn khác bao nhiêu % so với cộng đồng
- Điểm mạnh / yếu

---

## 4.3 Style Deviation

- Aggressive hơn hay conservative hơn cộng đồng
- Có đi ngược số đông không

---

## 4.4 Accuracy Benchmark

- Bạn vs Top 10%
- Bạn vs Average user

---

# 5. Gamification / Achievement Analytics

## 5.1 Badges

- Goal Prophet (đoán đúng tỷ số nhiều)
- Win Master (đoán đúng kết quả)
- Giant Killer (đoán cửa dưới đúng)
- Night Owl (predict ban đêm)

---

## 5.2 Progress Tracking

- Số lần đạt badge
- Tiến trình unlock

---

# 6. Advanced Metrics (Differentiator)

## 6.1 Consistency Score

- Mức độ ổn định dự đoán

---

## 6.2 Crowd Agreement Index

- % trùng với cộng đồng

---

## 6.3 Contrarian Index

- Mức độ đi ngược số đông

---

## 6.4 Score Precision

- Độ chính xác tỷ số chính xác

---

## 6.5 Prediction Evolution

- Thay đổi phong cách theo thời gian

---

# 7. UI/UX Recommendation

## Dashboard structure:

1. Overview
2. My Analytics
3. Community Analytics
4. Compare Me vs Community
5. Achievements

---

## Core charts to prioritize MVP:

- Accuracy Timeline
- Prediction Distribution
- Favorite Teams
- Community Trend
- Compare Table

---

# 8. Final Goal

Biến platform thành:

> “A prediction social analytics platform, not just a betting tool”

Người dùng quay lại không phải để đặt kèo, mà để xem:

- Mình dự đoán tốt hơn hay tệ hơn
- Mình thuộc kiểu người chơi nào
- Cộng đồng đang nghĩ gì
- Mình có đi ngược số đông không