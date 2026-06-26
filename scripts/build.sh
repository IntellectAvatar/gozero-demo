#!/usr/bin/env bash
#
# 跨平台打包脚本
# 用法: ./scripts/build.sh [版本号]
# 输出: build/ 目录下的各架构二进制文件
#

set -euo pipefail

VERSION="${1:-$(date +%Y%m%d%H%M%S)}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build/${VERSION}"

TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

declare -A PLATFORM_NAME
PLATFORM_NAME["linux/amd64"]="linux-amd64"
PLATFORM_NAME["linux/arm64"]="linux-arm64"
PLATFORM_NAME["windows/amd64"]="windows-amd64"

echo "============================================"
echo "  打包版本: ${VERSION}"
echo "  输出目录: ${BUILD_DIR}"
echo "============================================"

mkdir -p "${BUILD_DIR}"

build_service() {
    local service_name="$1"
    local src_dir="$2"
    local binary_name="$3"

    for target in "${TARGETS[@]}"; do
        GOOS="${target%/*}"
        GOARCH="${target#*/}"
        platform="${PLATFORM_NAME[$target]}"
        output="${BUILD_DIR}/${binary_name}-${platform}"

        if [ "${GOOS}" = "windows" ]; then
            output="${output}.exe"
        fi

        echo -n "  [${service_name}] ${platform} ... "

        CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
            go build -ldflags="-s -w -X main.version=${VERSION}" \
            -o "${output}" \
            "${src_dir}" 2>&1

        echo "✓ $(du -h "${output}" | cut -f1)"
    done
}

cd "${ROOT_DIR}"

echo ""
echo ">>> user-api (REST 服务)"
build_service "user-api" "./api/user" "user-api"

echo ""
echo ">>> user-rpc (gRPC 服务)"
build_service "user-rpc" "./rpc/user" "user-rpc"

echo ""
echo ">>> 复制配置文件"
mkdir -p "${BUILD_DIR}/config/api/user" "${BUILD_DIR}/config/rpc/user"
cp "${ROOT_DIR}/api/user/etc/user-api.yaml" "${BUILD_DIR}/config/api/user/"
cp "${ROOT_DIR}/rpc/user/etc/user-rpc.yaml" "${BUILD_DIR}/config/rpc/user/"

echo ""
echo "============================================"
echo "  打包完成! 输出目录: ${BUILD_DIR}"
echo "============================================"
ls -lh "${BUILD_DIR}/" | grep -v "^d\|^total\|config"
