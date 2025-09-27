import { useState, useCallback } from "react";

interface UseStreamOptions {
  onChunk?: (chunk: string) => void;
  onComplete?: (fullText: string) => void;
  onError?: (error: Error) => void;
}

export function useStream(options: UseStreamOptions = {}) {
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamedText, setStreamedText] = useState("");
  const [error, setError] = useState<Error | null>(null);

  const STREAM_ENDPOINT = "http://localhost:8080/api/gemini/resolve-conflicts";

  const startStream = useCallback(
    async (streamOptions?: { message?: string }) => {
      setIsStreaming(true);
      setStreamedText("");
      setError(null);

      try {
        const response = await fetch(STREAM_ENDPOINT, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Accept: "text/plain",
          },
          body: JSON.stringify(streamOptions),
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const reader = response.body?.getReader();
        if (!reader) {
          throw new Error("No response body reader available");
        }

        const decoder = new TextDecoder();
        let fullText = "";

        while (true) {
          const { done, value } = await reader.read();

          if (done) {
            break;
          }

          const chunk = decoder.decode(value, { stream: true });
          fullText += chunk;

          setStreamedText(fullText);
          options.onChunk?.(chunk);
        }

        options.onComplete?.(fullText);
      } catch (err) {
        const error =
          err instanceof Error ? err : new Error("Unknown error occurred");
        setError(error);
        options.onError?.(error);
      } finally {
        setIsStreaming(false);
      }
    },
    [options]
  );

  const reset = useCallback(() => {
    setStreamedText("");
    setError(null);
    setIsStreaming(false);
  }, []);

  return {
    isStreaming,
    streamedText,
    error,
    startStream,
    reset,
  };
}
