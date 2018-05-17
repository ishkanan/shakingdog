package data


func Max(x, y int) int {
  if x > y {
    return x
  }
  return y
}

func Left (s string, n int) string {
  if n < len(s) {
    return s[0:n]
  }
  return s
}
