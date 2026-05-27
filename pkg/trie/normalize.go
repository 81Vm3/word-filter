package trie

type NormFunc func(rune) rune

type NormalizeOptions struct {
	IgnoreCase  bool
	IgnoreWidth bool
}

func BuildNormalizer(opts NormalizeOptions) NormFunc {
	if !opts.IgnoreCase && !opts.IgnoreWidth {
		return identityRune
	}
	return func(r rune) rune {
		if opts.IgnoreWidth {
			r = foldWidth(r)
		}
		if opts.IgnoreCase {
			r = foldCase(r)
		}
		return r
	}
}

func identityRune(r rune) rune { return r }

func foldCase(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func foldWidth(r rune) rune {
	switch {
	case r == 0x3000:
		return ' '
	case r >= 0xFF01 && r <= 0xFF5E:
		return r - 0xFEE0
	default:
		return r
	}
}

func NormalizeString(s string, n NormFunc) string {
	if n == nil {
		return s
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		out = append(out, n(r))
	}
	return string(out)
}
