class Repodoctor < Formula
  desc "Static architecture analysis for software repositories"
  homepage "https://github.com/AdemFurkanATA/RepoDoctor"
  url "https://github.com/AdemFurkanATA/RepoDoctor/archive/refs/heads/main.tar.gz"
  version "main"
  sha256 :no_check
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = [
      "-s",
      "-w",
      "-X",
      "main.version=#{version}"
    ]
    system "go", "build", "-trimpath", *std_go_args(ldflags:), "."
  end

  test do
    assert_match "RepoDoctor", shell_output("#{bin}/repodoctor version")
  end
end
