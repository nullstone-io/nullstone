# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.100"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.100/nullstone_0.0.100_darwin_arm64.tar.gz"
      sha256 "2d48289187671088b75aefd3ac1ebf77da45c543f0b9068cddcf6b2b483093c7"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.100/nullstone_0.0.100_darwin_amd64.tar.gz"
      sha256 "87d564602d4b4e9e5743f67a6e3f36572608067adb38758c9658e021ffa69d98"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.100/nullstone_0.0.100_linux_arm64.tar.gz"
      sha256 "c00af4b41f2e5e6428f1ac38db5867dfd52e75b5fb11578be61dcefe8a477ff1"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.100/nullstone_0.0.100_linux_amd64.tar.gz"
      sha256 "92eca1906c138b081200aa3b30794b934743852f6f92b035ea59a5149ae71062"

      def install
        bin.install "nullstone"
      end
    end
  end
end
