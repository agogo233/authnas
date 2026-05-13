#!/bin/bash

# 前后端响应一致性全面测试脚本

BASE_URL="http://localhost:8080"
STATIC_DIR="/workspace/go-server/static"
PASS=0
FAIL=0

test_result() {
    local name="$1"
    local expected="$2"
    local actual="$3"
    if [[ "$expected" == "$actual" ]]; then
        echo "  [PASS] $name"
        ((PASS++))
    else
        echo "  [FAIL] $name - Expected: $expected, Got: $actual"
        ((FAIL++))
    fi
}

test_header() {
    local name="$1"
    local url="$2"
    local header="$3"
    local expected="$4"
    local actual=$(curl -s -I "$url" | grep -i "^$header:" | cut -d' ' -f2 | tr -d '\r\n')
    test_result "$name" "$expected" "$actual"
}

echo "=========================================="
echo "前后端响应一致性全面测试"
echo "=========================================="
echo ""

# 启动服务器
echo "[1] 启动Go服务器..."
cd /workspace/go-server && ./go-server-test &
SERVER_PID=$!
sleep 2

# 检查服务器是否启动
if ! curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" | grep -q "200"; then
    echo "服务器启动失败!"
    exit 1
fi
echo "服务器已启动 (PID: $SERVER_PID)"
echo ""

# 1. 基础HTTP状态码测试
echo "[2] 基础HTTP状态码测试"
echo "---------------------------"
test_result "/" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/)" "200"
test_result "/favicon.svg" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/favicon.svg)" "200"
test_result "/icons.svg" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/icons.svg)" "200"
test_result "/index.html" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/index.html)" "200"
test_result "/api/health" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/api/health)" "200"
test_result "/404-test" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/nonexistent-page)" "200"
echo ""

# 2. MIME类型测试
echo "[3] MIME类型测试"
echo "---------------------------"
test_header "index.html" "$BASE_URL/" "content-type" "text/html"
test_header "favicon.svg" "$BASE_URL/favicon.svg" "content-type" "image/svg+xml"
test_header "icons.svg" "$BASE_URL/icons.svg" "content-type" "image/svg+xml"

JS_FILE=$(ls $STATIC_DIR/assets/*.js 2>/dev/null | head -1 | xargs basename)
if [[ -n "$JS_FILE" ]]; then
    test_header "JS file" "$BASE_URL/assets/$JS_FILE" "content-type" "application/javascript"
fi

CSS_FILE=$(ls $STATIC_DIR/assets/*.css 2>/dev/null | head -1 | xargs basename)
if [[ -n "$CSS_FILE" ]]; then
    test_header "CSS file" "$BASE_URL/assets/$CSS_FILE" "content-type" "text/css"
fi
echo ""

# 3. HTML中的资源引用验证
echo "[4] HTML资源引用验证"
echo "---------------------------"
INDEX_HTML=$(curl -s $BASE_URL/)
ASSETS_COUNT=$(echo "$INDEX_HTML" | grep -o 'src="/assets/[^"]*' | wc -l)
test_result "index.html has asset references" "$ASSETS_COUNT" -gt 0

ASSET_PATHS=$(echo "$INDEX_HTML" | grep -o 'src="/assets/[^"]*' | sed 's/src="//;s/"$//')
while IFS= read -r path; do
    if [[ -n "$path" ]]; then
        status=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL$path")
        test_result "HTML references: $path" "$status" "200"
    fi
done <<< "$ASSET_PATHS"
echo ""

# 4. 所有静态文件可访问性
echo "[5] 所有静态文件可访问性测试"
echo "---------------------------"
for file in $(find $STATIC_DIR -type f); do
    rel_path=${file#$STATIC_DIR}
    status=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL$rel_path")
    test_result "Static file: $rel_path" "$status" "200"
done
echo ""

# 5. Content-Length验证
echo "[6] Content-Length验证"
echo "---------------------------"
for file in $(find $STATIC_DIR -type f -name "*.html" -o -name "*.svg" -o -name "*.js" -o -name "*.css" | head -10); do
    rel_path=${file#$STATIC_DIR}
    orig_size=$(stat -c%s "$file")
    curl_size=$(curl -s -I "$BASE_URL$rel_path" | grep -i "^content-length:" | awk '{print $2}' | tr -d '\r')
    test_result "Size match: $rel_path" "$orig_size" "$curl_size"
done
echo ""

# 6. 缓存头检查
echo "[7] 缓存头检查"
echo "---------------------------"
CACHE_HEADERS=("cache-control" "etag" "expires" "last-modified" "vary")
for header in "${CACHE_HEADERS[@]}"; do
    result=$(curl -s -I "$BASE_URL/index.html" | grep -i "^$header:" || echo "NOT_FOUND")
    echo "  index.html $header: $(echo $result | cut -d' ' -f2)"
done
echo ""

# 7. CORS头检查
echo "[8] CORS头检查"
echo "---------------------------"
CORS_HEADERS=("access-control-allow-origin" "access-control-allow-methods" "access-control-allow-headers")
for header in "${CORS_HEADERS[@]}"; do
    result=$(curl -s -I -X OPTIONS -H "Origin: http://example.com" "$BASE_URL/api/health" 2>/dev/null | grep -i "^$header:" || echo "NOT_FOUND")
    echo "  API CORS $header: $(echo $result | cut -d' ' -f2)"
done
echo ""

# 8. API端点测试
echo "[9] API端点基础测试"
echo "---------------------------"
test_result "/api/health" "$(curl -s $BASE_URL/api/health)" '{"status":"ok"}'
test_result "/api/config" "$(curl -s -o /dev/null -w '%{http_code}' $BASE_URL/api/config)" "200"
echo ""

# 9. SPA路由测试
echo "[10] SPA路由测试"
echo "---------------------------"
SPA_ROUTES=("/login" "/register" "/dashboard" "/admin" "/settings" "/profile")
for route in "${SPA_ROUTES[@]}"; do
    content=$(curl -s "$BASE_URL$route")
    has_index=$(echo "$content" | grep -c "index.html" || true)
    test_result "SPA route: $route returns index.html" "$has_index" 1
done
echo ""

# 10. 安全头检查
echo "[11] 安全头检查"
echo "---------------------------"
SECURITY_HEADERS=("x-content-type-options" "x-frame-options" "x-xss-protection")
for header in "${SECURITY_HEADERS[@]}"; do
    result=$(curl -s -I "$BASE_URL/" | grep -i "^$header:" || echo "NOT_FOUND")
    echo "  $header: $(echo $result | cut -d' ' -f2)"
done
echo ""

# 11. Gzip/压缩支持检查
echo "[12] 压缩支持检查"
echo "---------------------------"
ACCEPT_ENCODING_TEST=$(curl -s -I -H "Accept-Encoding: gzip" "$BASE_URL/assets/$JS_FILE" | grep -i "^content-encoding:" || echo "NOT_FOUND")
echo "  JS gzipped: $ACCEPT_ENCODING_TEST"
echo ""

# 清理
kill $SERVER_PID 2>/dev/null

echo "=========================================="
echo "测试完成"
echo "通过: $PASS"
echo "失败: $FAIL"
echo "=========================================="

if [[ $FAIL -gt 0 ]]; then
    exit 1
fi
exit 0