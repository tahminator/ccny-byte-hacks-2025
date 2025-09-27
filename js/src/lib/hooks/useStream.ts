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

  const startStream = useCallback(
    async (url: string, streamOptions?: { message?: string }) => {
      setIsStreaming(true);
      setStreamedText("");
      setError(null);

      try {
        // Build URL with query parameters for GET request
        const urlWithParams = new URL(url);
        if (streamOptions?.message) {
          urlWithParams.searchParams.set("message", streamOptions.message);
        }

        const response = await fetch(urlWithParams.toString(), {
          method: "GET",
          headers: {
            Accept: "text/plain",
          },
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
