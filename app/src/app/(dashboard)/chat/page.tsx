"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Send, ChevronDown, ChevronUp, MessageSquare, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { queryRAG, listHubDocuments, type HubDocument, type QueryResponse } from "@/lib/api";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  sources?: QueryResponse["sources"];
  loading?: boolean;
}

const EXAMPLE_PROMPTS = [
  "Hợp đồng này có những điều khoản quan trọng nào?",
  "Tổng số tiền trong hóa đơn là bao nhiêu?",
  "CV này có những kỹ năng gì nổi bật?",
  "Tóm tắt nội dung chính của tài liệu",
];

function SourceItem({ source }: { source: QueryResponse["sources"][number] }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-md border border-gray-100 bg-gray-50 text-xs">
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-gray-100"
      >
        <span className="text-indigo-600 font-medium truncate">
          Doc: {source.doc_id.slice(0, 8)}…
        </span>
        <div className="flex items-center gap-2 shrink-0">
          <span className="text-gray-400">
            Điểm: {(source.score * 100).toFixed(1)}%
          </span>
          {open ? (
            <ChevronUp className="h-3 w-3 text-gray-400" />
          ) : (
            <ChevronDown className="h-3 w-3 text-gray-400" />
          )}
        </div>
      </button>
      {open && (
        <div className="border-t border-gray-100 px-3 py-2 text-gray-600 whitespace-pre-wrap">
          {source.chunk_text}
        </div>
      )}
    </div>
  );
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [docs, setDocs] = useState<HubDocument[]>([]);
  const [selectedDocIds, setSelectedDocIds] = useState<string[]>([]);
  const bottomRef = useRef<HTMLDivElement>(null);

  // Load indexed documents for filter
  useEffect(() => {
    listHubDocuments()
      .then((list) => setDocs(list.filter((d) => d.status === "indexed")))
      .catch(() => {});
  }, []);

  // Scroll to bottom on new message
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const sendMessage = useCallback(async () => {
    const question = input.trim();
    if (!question || isLoading) return;

    const userMsg: Message = {
      id: crypto.randomUUID(),
      role: "user",
      content: question,
    };
    const placeholderId = crypto.randomUUID();
    const placeholder: Message = {
      id: placeholderId,
      role: "assistant",
      content: "",
      loading: true,
    };

    setMessages((prev) => [...prev, userMsg, placeholder]);
    setInput("");
    setIsLoading(true);

    try {
      const res = await queryRAG({
        question,
        doc_ids: selectedDocIds.length > 0 ? selectedDocIds : undefined,
      });

      setMessages((prev) =>
        prev.map((m) =>
          m.id === placeholderId
            ? { ...m, content: res.answer, sources: res.sources, loading: false }
            : m
        )
      );
    } catch (err) {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === placeholderId
            ? {
                ...m,
                content:
                  "Lỗi: " +
                  (err instanceof Error ? err.message : "Không thể kết nối đến server"),
                loading: false,
              }
            : m
        )
      );
    } finally {
      setIsLoading(false);
    }
  }, [input, isLoading, selectedDocIds]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const toggleDocFilter = (id: string) => {
    setSelectedDocIds((prev) =>
      prev.includes(id) ? prev.filter((d) => d !== id) : [...prev, id]
    );
  };

  return (
    <div className="flex h-[calc(100vh-8rem)] flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Chat AI</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Đặt câu hỏi về tài liệu của bạn
          </p>
        </div>
        {docs.length > 0 && (
          <div className="flex items-center gap-2 flex-wrap max-w-md justify-end">
            <span className="text-xs text-muted-foreground">Lọc tài liệu:</span>
            {docs.map((doc) => (
              <button
                key={doc.id}
                onClick={() => toggleDocFilter(doc.id)}
                className={`rounded-full px-2 py-0.5 text-xs border transition-colors ${
                  selectedDocIds.includes(doc.id)
                    ? "bg-indigo-600 text-white border-indigo-600"
                    : "bg-white text-gray-600 border-gray-200 hover:border-gray-400"
                }`}
                title={doc.original_filename}
              >
                {doc.original_filename.length > 20
                  ? doc.original_filename.slice(0, 20) + "…"
                  : doc.original_filename}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Chat area */}
      <Card className="flex flex-1 flex-col overflow-hidden">
        <CardContent className="flex flex-1 flex-col overflow-hidden p-0">
          {/* Messages */}
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {messages.length === 0 ? (
              <div className="flex h-full flex-col items-center justify-center gap-6 py-12 text-center">
                <MessageSquare className="h-12 w-12 text-gray-200" />
                <div>
                  <p className="text-lg font-medium text-gray-500">
                    Bắt đầu đặt câu hỏi về tài liệu của bạn
                  </p>
                  <p className="text-sm text-gray-400 mt-1">
                    Hệ thống sẽ tìm kiếm thông tin từ các tài liệu đã được index
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-2 max-w-lg w-full">
                  {EXAMPLE_PROMPTS.map((prompt) => (
                    <button
                      key={prompt}
                      onClick={() => setInput(prompt)}
                      className="rounded-lg border border-gray-200 p-3 text-left text-sm text-gray-600 hover:bg-gray-50 hover:border-gray-300 transition-colors"
                    >
                      {prompt}
                    </button>
                  ))}
                </div>
              </div>
            ) : (
              messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
                >
                  <div
                    className={`max-w-[75%] rounded-2xl px-4 py-3 text-sm ${
                      msg.role === "user"
                        ? "bg-indigo-600 text-white"
                        : "bg-gray-100 text-gray-800"
                    }`}
                  >
                    {msg.loading ? (
                      <div
                        className="flex items-center gap-2 text-gray-500"
                        aria-label="Đang tải câu trả lời"
                      >
                        <Loader2 className="h-4 w-4 animate-spin" />
                        <span>Đang xử lý...</span>
                      </div>
                    ) : (
                      <div className="whitespace-pre-wrap">{msg.content}</div>
                    )}

                    {/* Sources */}
                    {!msg.loading &&
                      msg.sources &&
                      msg.sources.length > 0 && (
                        <div className="mt-3 space-y-1">
                          <p className="text-xs font-medium text-gray-500 mb-1">
                            Nguồn tham khảo:
                          </p>
                          {msg.sources.map((src, i) => (
                            <SourceItem key={i} source={src} />
                          ))}
                        </div>
                      )}
                  </div>
                </div>
              ))
            )}
            <div ref={bottomRef} />
          </div>

          {/* Input area */}
          <div className="border-t p-4">
            <div className="flex items-end gap-2">
              <textarea
                aria-label="Nhập câu hỏi"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Đặt câu hỏi về tài liệu... (Enter để gửi, Shift+Enter để xuống dòng)"
                rows={2}
                className="flex-1 resize-none rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
              />
              <Button
                onClick={sendMessage}
                disabled={!input.trim() || isLoading}
                className="shrink-0"
                aria-label="Gửi tin nhắn"
              >
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Send className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
