#!/usr/bin/env bash
#
# 安装脚本 — 将二进制和 systemd 服务部署到目标服务器
# 用法: sudo ./scripts/install.sh [version]
# 示例: sudo ./scripts/install.sh 20260624170000
#

set -euo pipefail

VERSION="${1:-$(ls -1 build/ | sort -r | head -1)}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"
BUILD_DIR="${ROOT_DIR}/build/${VERSION}"
INSTALL_DIR="/opt/gozero-demo"

# 检测系统架构
ARCH=$(uname -m)
case "${ARCH}" in
    x86_64)  PLATFORM="linux-amd64" ;;
    aarch64) PLATFORM="linux-arm64" ;;
    *) echo "不支持的架构: ${ARCH}"; exit 1 ;;
esac

echo "============================================"
echo "  安装 gozero-demo 微服务"
echo "  版本: ${VERSION}"
echo "  架构: ${ARCH} (${PLATFORM})"
echo "  安装目录: ${INSTALL_DIR}"
echo "============================================"

# 检查 root
if [ "$EUID" -ne 0 ]; then
    echo "请用 sudo 运行此脚本"
    exit 1
fi

# 确认二进制存在
HELLO_BIN="${BUILD_DIR}/hello-api-${PLATFORM}"
USER_BIN="${BUILD_DIR}/user-rpc-${PLATFORM}"
if [ ! -f "${HELLO_BIN}" ] || [ ! -f "${USER_BIN}" ]; then
    echo "错误: 找不到 ${PLATFORM} 架构的二进制文件"
    echo "  需要: ${HELLO_BIN}"
    echo "  需要: ${USER_BIN}"
    echo ""
    echo "请先运行构建脚本: ./scripts/build.sh"
    exit 1
fi

# 创建安装目录
echo ">>> 创建目录结构"
mkdir -p "${INSTALL_DIR}"/{logs,bin,config/api/hello,config/rpc/user}

# 复制二进制
echo ">>> 安装二进制文件"
cp "${HELLO_BIN}" "${INSTALL_DIR}/hello-api"
cp "${USER_BIN}" "${INSTALL_DIR}/user-rpc"
chmod 755 "${INSTALL_DIR}/hello-api" "${INSTALL_DIR}/user-rpc"

# 复制配置文件（保留已有配置）
echo ">>> 安装配置文件"
cp -n "${BUILD_DIR}/config/api/hello/hello-api.yaml" "${INSTALL_DIR}/config/api/hello/" 2>/dev/null || true
cp -n "${BUILD_DIR}/config/rpc/user/user-rpc.yaml" "${INSTALL_DIR}/config/rpc/user/" 2>/dev/null || true

# 创建日志目录
touch "${INSTALL_DIR}/logs/.keep"

# 安装 systemd 服务
echo ">>> 安装 systemd 服务"
cp "${ROOT_DIR}/scripts/hello-api.service" /etc/systemd/system/
cp "${ROOT_DIR}/scripts/user-rpc.service" /etc/systemd/system/
systemctl daemon-reload

echo ""
echo "============================================"
echo "  安装完成!"
echo ""
echo "  启动服务:"
echo "    sudo systemctl start user-rpc"
echo "    sudo systemctl start hello-api"
echo ""
echo "  设置开机自启:"
echo "    sudo systemctl enable user-rpc hello-api"
echo ""
echo "  查看状态:"
echo "    sudo systemctl status user-rpc hello-api"
echo ""
echo "  查看日志:"
echo "    sudo journalctl -u user-rpc -f"
echo "    sudo journalctl -u hello-api -f"
echo "============================================"
