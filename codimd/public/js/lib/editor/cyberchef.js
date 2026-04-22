// Pure-JS CyberChef-style transforms. Each takes a string, returns a string,
// and throws Error on invalid input. No network, no deps.

export function b64Encode (s) {
  return btoa(unescape(encodeURIComponent(s)))
}

export function b64Decode (s) {
  return decodeURIComponent(escape(atob(s.trim())))
}

export function urlEncode (s) {
  return encodeURIComponent(s)
}

export function urlDecode (s) {
  return decodeURIComponent(s)
}

export function hexEncode (s) {
  return Array.from(new TextEncoder().encode(s.trim()))
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
}

export function hexDecode (s) {
  const clean = s.replace(/\s+/g, '')
  if (clean.length % 2 !== 0 || !/^[0-9a-fA-F]*$/.test(clean)) {
    throw new Error('Invalid hex input')
  }
  const bytes = new Uint8Array(clean.length / 2)
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(clean.substr(i * 2, 2), 16)
  }
  return new TextDecoder().decode(bytes)
}

export function htmlEncode (s) {
  return s.replace(/[&<>"']/g, c => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
  }[c]))
}

export function htmlDecode (s) {
  const doc = new DOMParser().parseFromString(s, 'text/html')
  return doc.documentElement.textContent
}

export function jsonBeautify (s) {
  return JSON.stringify(JSON.parse(s), null, 2)
}

export function jwtDecode (s) {
  const parts = s.trim().split('.')
  if (parts.length !== 3) throw new Error('Not a JWT (expected 3 dot-separated parts)')
  const b64urlDecode = (x) => atob(x.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(x.length / 4) * 4, '='))
  const header = JSON.parse(b64urlDecode(parts[0]))
  const payload = JSON.parse(b64urlDecode(parts[1]))
  return JSON.stringify({ header, payload, signature: parts[2] }, null, 2)
}
