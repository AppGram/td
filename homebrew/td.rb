# Homebrew formula for td
# This file should be placed in appgram/homebrew-tap repo under Formula/td.rb

class Td < Formula
  desc "Minimal todo app with beautiful TUI, inline task syntax, and smart dashboard"
  homepage "https://github.com/appgram/td"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/appgram/td/releases/download/v#{version}/td_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    else
      url "https://github.com/appgram/td/releases/download/v#{version}/td_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/appgram/td/releases/download/v#{version}/td_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/appgram/td/releases/download/v#{version}/td_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "td"
  end

  def caveats
    <<~EOS
      Data is stored in: ~/.config/td/td.db

      Quick start:
        td              # Launch TUI
        td -a "task"    # Add task from CLI

      Inline syntax when adding tasks:
        td -a "Buy milk #shopping @tomorrow !high"
    EOS
  end

  test do
    assert_match "td", shell_output("#{bin}/td -version 2>&1", 0)
  end
end
