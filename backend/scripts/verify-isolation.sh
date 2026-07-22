#!/usr/bin/env bash
# Verifies that two accounts have completely isolated carts and wishlists.
#
# Run against a running backend (migrations applied, catalogue seeded):
#   ./scripts/verify-isolation.sh [BASE_URL]
# Default BASE_URL is http://localhost:8080/api/v1
#
# It creates two fresh users, has each store cart items and saved collections,
# and then asserts that neither can see or delete the other's data.

set -uo pipefail

BASE="${1:-http://localhost:8080/api/v1}"
STAMP="$(date +%s)$RANDOM"
A_EMAIL="iso-a-${STAMP}@example.com"
B_EMAIL="iso-b-${STAMP}@example.com"
PASS="correct-horse-battery"

pass=0
fail=0
ok()   { echo "  PASS  $1"; pass=$((pass+1)); }
bad()  { echo "  FAIL  $1"; fail=$((fail+1)); }

jqq() { python3 -c "import sys,json;d=json.load(sys.stdin);print(eval('d$1'))" 2>/dev/null; }

register() {
  curl -s -X POST "$BASE/auth/register" -H 'Content-Type: application/json' \
    -d "{\"name\":\"$1\",\"email\":\"$2\",\"password\":\"$PASS\"}"
}

echo "Base URL: $BASE"
echo

# ── Create two users ────────────────────────────────────────────────────────
TOKEN_A=$(register "Iso A" "$A_EMAIL" | jqq "['data']['tokens']['accessToken']")
TOKEN_B=$(register "Iso B" "$B_EMAIL" | jqq "['data']['tokens']['accessToken']")

if [ -z "${TOKEN_A:-}" ] || [ -z "${TOKEN_B:-}" ]; then
  echo "Could not create both users — is the backend running at $BASE?"
  exit 1
fi
echo "Created two users."
echo

authA() { curl -s -H "Authorization: Bearer $TOKEN_A" -H 'Content-Type: application/json' "$@"; }
authB() { curl -s -H "Authorization: Bearer $TOKEN_B" -H 'Content-Type: application/json' "$@"; }

# ── Pick two real products from the catalogue ───────────────────────────────
PRODUCTS=$(authA "$BASE/products?limit=2")
P1=$(echo "$PRODUCTS" | jqq "['data'][0]['id']")
P2=$(echo "$PRODUCTS" | jqq "['data'][1]['id']")
if [ -z "${P1:-}" ]; then
  echo "No products returned — seed the catalogue first."
  exit 1
fi

# ── User A fills their cart, wishlist and saved collections ────────────────
authA -X POST "$BASE/cart" -d "{\"productId\":\"$P1\"}" > /dev/null
authA -X POST "$BASE/favorites" -d "{\"productId\":\"$P1\"}" > /dev/null
SAVED_A=$(authA -X POST "$BASE/saved-collections" \
  -d "{\"name\":\"A Look\",\"department\":\"Women\",\"prompt\":\"a\",\"collection\":{\"title\":\"A Look\",\"department\":\"Women\",\"items\":[],\"totalPrice\":0}}")
SAVED_A_ID=$(echo "$SAVED_A" | jqq "['data']['id']")

# ── User B stores something different ──────────────────────────────────────
authB -X POST "$BASE/cart" -d "{\"productId\":\"${P2:-$P1}\"}" > /dev/null

echo "Isolation assertions:"

# 1. A's cart contains exactly A's product
A_CART_N=$(authA "$BASE/cart" | jqq "['data'].__len__()")
A_CART_P=$(authA "$BASE/cart" | jqq "['data'][0]['product']['id']")
[ "$A_CART_N" = "1" ] && [ "$A_CART_P" = "$P1" ] \
  && ok "A's cart holds only A's item" || bad "A's cart wrong (n=$A_CART_N)"

# 2. B's cart does NOT contain A's item
B_CART_N=$(authB "$BASE/cart" | jqq "['data'].__len__()")
B_HAS_A=$(authB "$BASE/cart" | grep -c "$P1" || true)
[ "$B_CART_N" = "1" ] && ok "B's cart holds only B's item" || bad "B's cart wrong (n=$B_CART_N)"
if [ "${P2:-}" != "" ] && [ "$P2" != "$P1" ]; then
  [ "$B_HAS_A" = "0" ] && ok "B cannot see A's cart item" || bad "LEAK: B sees A's cart item"
fi

# 3. B's wishlist is empty (A favourited, B did not)
B_FAV_N=$(authB "$BASE/favorites" | jqq "['data'].__len__()")
[ "$B_FAV_N" = "0" ] && ok "B's wishlist is empty" || bad "LEAK: B's wishlist has $B_FAV_N item(s)"

# 4. B's saved collections are empty
B_SAVED_N=$(authB "$BASE/saved-collections" | jqq "['data'].__len__()")
[ "$B_SAVED_N" = "0" ] && ok "B's saved collections empty" || bad "LEAK: B sees $B_SAVED_N saved collection(s)"

# 5. B cannot delete A's saved collection even with the exact id
if [ -n "${SAVED_A_ID:-}" ]; then
  CODE=$(curl -s -o /dev/null -w '%{http_code}' -X DELETE \
    -H "Authorization: Bearer $TOKEN_B" "$BASE/saved-collections/$SAVED_A_ID")
  [ "$CODE" = "404" ] && ok "B cannot delete A's collection (404)" \
    || bad "OWNERSHIP HOLE: B deleting A's collection returned $CODE"
  A_STILL=$(authA "$BASE/saved-collections" | jqq "['data'].__len__()")
  [ "$A_STILL" = "1" ] && ok "A's collection survived B's attempt" || bad "A's collection was destroyed by B"
fi

# 6. B cannot remove A's cart item
CODE=$(curl -s -o /dev/null -w '%{http_code}' -X DELETE \
  -H "Authorization: Bearer $TOKEN_B" "$BASE/cart/$P1")
[ "$CODE" = "404" ] && ok "B cannot delete A's cart item (404)" \
  || bad "OWNERSHIP HOLE: B deleting A's cart item returned $CODE"

echo
echo "  $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
