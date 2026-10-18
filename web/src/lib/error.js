// Returns the API error message from an axios error, or the fallback
export function errorText(err, fallback) {
  return err.response?.data?.errorMessage || fallback
}
