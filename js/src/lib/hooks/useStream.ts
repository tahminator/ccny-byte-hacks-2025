import { useState, useCallback } from "react";

// Content from testRepo/main.go for testing
const TEST_CONFLICT_CONTENT = `package testrepo

import (
	"fmt"
	"log"
	"os"
)

type User struct {
	ID       int    \`json:"id"\`
	Name     string \`json:"name"\`
	Email    string \`json:"email"\`
	Password string \`json:"-"\`
}

<<<<<<< HEAD
func (u *User) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}
=======
func (u *User) Validate() error {
	if len(u.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if u.Email == "" || !strings.Contains(u.Email, "@") {
		return fmt.Errorf("email must be valid")
	}
	return nil
}
>>>>>>> feature/user-validation

<<<<<<< HEAD
func main() {
	user := &User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	if err := user.Validate(); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("User created successfully")
}
=======
func main() {
	user := &User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	if err := user.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return
	}
	
	fmt.Printf("User %s created successfully\\n", user.Name)
}
>>>>>>> feature/improved-logging

<<<<<<< HEAD
func GetUserByID(id int) (*User, error) {
	// TODO: implement database lookup
	return nil, fmt.Errorf("not implemented")
}
=======
func GetUserByID(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	// TODO: implement database lookup
	return nil, fmt.Errorf("not implemented")
}
>>>>>>> feature/input-validation`;

interface UseStreamOptions {
  onChunk?: (chunk: string) => void;
  onComplete?: (fullText: string) => void;
  onError?: (error: Error) => void;
}

export function useStream(options: UseStreamOptions = {}) {
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamedText, setStreamedText] = useState("");
  const [error, setError] = useState<Error | null>(null);

  const STREAM_ENDPOINT = "/api/gemini/resolve-conflicts-file-stream";

  const startStream = useCallback(
    async (streamOptions?: {
      conflictContent?: string;
      filePath?: string;
      userQuery?: string;
      repoHash?: string;
    }) => {
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
          body: JSON.stringify({
            conflict_content:
              streamOptions?.conflictContent || TEST_CONFLICT_CONTENT, // replace this with actual files
            file_path: streamOptions?.filePath || "",
            user_query: streamOptions?.userQuery || "",
            repo_hash: streamOptions?.repoHash || "",
          }),
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
