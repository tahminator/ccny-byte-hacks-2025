type ApiSuccess<T> = {
  success: true;
  message: string;
  payload: T;
};

type ApiError = {
  success: false;
  message: string;
  /**
   * You can confirm if payload is of type `T`
   * by checking if success is true,
   */
  payload: undefined;
};

/**
 * Generic wrapper for all responses returned by Codebloom's backend.
 * true
 * true
 * true
 *
 * @note - The `payload` of type `T` attribute is only available in the SuccessType, not ErrorType.
 * You can attempt to access `payload` by checking if `success` is true.
 *
 * @example
 * ```ts
 * const json = (await response.json()) as ApiResponse<Todo>;
 *
 * json.payload; // Property 'payload' does not exist on type 'ApiResponse<Todo>'.
 *
 * if (!json.success) {
 *  return;
 * }
 *
 * const todo = json.payload; // no error, type of Todo.
 * ```
 */
export type ApiResponder<T> = ApiSuccess<T> | ApiError;
