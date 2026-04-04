user upload/capture single/multi images => ai scan => user review => user click save each/all
LLM vision theo mình đánh giá tốt nhất là gemini, còn nhỏ nhẹ tiếng việt ngon thì vintern-1b cũng xịn vãi rồi, ngoài ra nên có phase pre-processing input image ngon nghẻ trc khi gửi cho AI

Ý em là mặc dù nhiều model LLM bây giờ xử lý vision khá tốt, nhưng cũng không vì thế mà mình mặc định đẩy hết cho LLM xử lý (gửi trực tiếp cho LLM).
Vì hình hóa đơn mỗi ng chụp một kiểu, với case đẹp ko sao, còn những case xấu ví dụ: bị méo mó, ngược sáng, bị mờ, ... thì mình nên có phase tiền xử lý trc để chuẩn hóa nó (ví dụ: chỉnh cho thẳng, cắt gọt lấy mỗi hình hóa đơn, ...) trước khi gửi cho LLM thì đảm bảo, kết quả sẽ tốt và chính xác hơn nữa.
