# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.23"
  license "MIT"
  bottle :unneeded

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.23/nullstone_0.0.23_Darwin_x86_64.tar.gz"
      sha256 "0bfa5295edf0fc9c8370adfba47b6ab491d7c2cbec1ce8cff826e9b976d806b5"
    end
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.23/nullstone_0.0.23_Darwin_arm64.tar.gz"
      sha256 "64153159705de25bee7c9c31cc5b829e55a6eea0a8703ad292fd28c1d26d382e"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.23/nullstone_0.0.23_Linux_x86_64.tar.gz"
      sha256 "536f74f0221a662e51d48906bf2410090b63ecf5275446d58000e6890f6b417f"
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.23/nullstone_0.0.23_Linux_arm64.tar.gz"
      sha256 "87c55b22aaf9ad6ff2c39247fc53b34dcb586794e1603216c29af43a602f8ece"
    end
  end

  depends_on "go"

  def install
    bin.install "nullstone"
  end
end
