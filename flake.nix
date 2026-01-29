{
  description = "TubeViz - audio visualiser that renders MP4 video from podcast audio";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    let
      supportedSystems = [ "x86_64-linux" ]; # ffmpeg-statigo ships linux_amd64 static libs
    in
    flake-utils.lib.eachSystem supportedSystems (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or "dev";
        ffmpegStatigoLib = pkgs.fetchurl {
          url = "https://github.com/ChrisLAS/ffmpeg-statigo-mirror/releases/download/lib-8.0.1.0/ffmpeg-linux-amd64.tar.gz";
          hash = "sha256-nfakHmdmb++pR7ridXo2QN/V47p9KsVwfI0uo71Pkbk=";
        };
        ffmpegStatigoSrc = pkgs.fetchFromGitHub {
          owner = "ChrisLAS";
          repo = "ffmpeg-statigo-mirror";
          rev = "4b9e3969d5e2d7140db5743b9c42b525110920bf";
              hash = "sha256-xkuvggFPnKv8d/0A7yCuGTw/48MAOCabHFMW6i920Rs=";
        };
      in
      {
        packages.tubeviz = pkgs.buildGoModule {
          pname = "tubeviz";
          inherit version;
          src = self;
          subPackages = [ "cmd/jivefire" ];
          # Update with `nix build .#tubeviz` when deps change.
          vendorHash = "sha256-ZQi5efJlY9fNRDtR8oA2Fm/bM2MyG7nIk7uypv2dzC8=";
          ldflags = [ "-X main.version=${version}" ];
          nativeBuildInputs = [ pkgs.pkg-config pkgs.gnutar ];
          doCheck = false;
          postPatch = ''
            rm -rf third_party/ffmpeg-statigo
            mkdir -p third_party
            cp -R ${ffmpegStatigoSrc} third_party/ffmpeg-statigo
            chmod -R u+w third_party/ffmpeg-statigo
            mkdir -p third_party/ffmpeg-statigo/lib
            tar -xzf ${ffmpegStatigoLib} -C third_party/ffmpeg-statigo/lib
          '';
          preBuild = ''
            if [ ! -d vendor ]; then
              echo "vendor directory missing; expected buildGoModule vendor phase"
              exit 1
            fi
            chmod -R u+w vendor

            mkdir -p vendor/github.com/linuxmatters/ffmpeg-statigo/lib/linux_amd64
            cp third_party/ffmpeg-statigo/lib/linux_amd64/libffmpeg.a \
              vendor/github.com/linuxmatters/ffmpeg-statigo/lib/linux_amd64/

            rm -rf vendor/github.com/linuxmatters/ffmpeg-statigo/include
            cp -R third_party/ffmpeg-statigo/include \
              vendor/github.com/linuxmatters/ffmpeg-statigo/include
          '';
          postInstall = ''
            mv "$out/bin/jivefire" "$out/bin/tubeviz"
          '';
          meta.mainProgram = "tubeviz";
        };

        packages.default = self.packages.${system}.tubeviz;

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.tubeviz;
          exePath = "/bin/tubeviz";
        };

        devShells.default = pkgs.mkShell {
          packages =
            with pkgs;
            [
              curl
              ffmpeg-full
              gnugrep
              gcc
              go
              just
              pciutils
              vhs
            ]
            ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
              vulkan-loader # Required for Vulkan accelerated encoders on Linux
              intel-media-driver # VA-API driver for Intel GPUs (iHD_drv_video.so)
              vpl-gpu-rt # oneVPL runtime for Intel GPUs (QSV backend)
            ];

          # Make GPU drivers visible for hardware-accelerated encoding
          # Linux: NixOS mounts GPU drivers under /run/opengl-driver/lib
          shellHook = pkgs.lib.optionalString pkgs.stdenv.isLinux ''
            # If the opengl driver directory exists, prepend it to LD_LIBRARY_PATH
            if [ -d "/run/opengl-driver/lib" ]; then
              if [ -z "$LD_LIBRARY_PATH" ]; then
                export LD_LIBRARY_PATH="/run/opengl-driver/lib"
              else
                export LD_LIBRARY_PATH="/run/opengl-driver/lib:$LD_LIBRARY_PATH"
              fi
            fi
            # Add vulkan-loader library path for h264_vulkan encoder
            export LD_LIBRARY_PATH="${pkgs.vulkan-loader}/lib:$LD_LIBRARY_PATH"
            # Add Intel media driver, and oneVPL libraries for QSV
            export LD_LIBRARY_PATH="${pkgs.intel-media-driver}/lib:$LD_LIBRARY_PATH"
            export LD_LIBRARY_PATH="${pkgs.vpl-gpu-rt}/lib:$LD_LIBRARY_PATH"
            # oneVPL runtime search path for QSV (11th gen+ Intel only)
            export ONEVPL_SEARCH_PATH="${pkgs.vpl-gpu-rt}/lib"

            # VA-API driver discovery for libva
            # Use system drivers if available, fall back to nix package for Intel
            if [ -d "/run/opengl-driver/lib/dri" ]; then
              export LIBVA_DRIVERS_PATH="/run/opengl-driver/lib/dri"
            fi
            # Auto-detect VA-API driver based on GPU vendor (prefer Intel for VA-API)
            if lspci -d ::0300 2>/dev/null | grep -qi intel; then
              export LIBVA_DRIVER_NAME="iHD"
              # Ensure Intel driver path is set even without system drivers
              export LIBVA_DRIVERS_PATH="${pkgs.intel-media-driver}/lib/dri:''${LIBVA_DRIVERS_PATH:-}"
            elif lspci -d ::0300 2>/dev/null | grep -qi amd; then
              export LIBVA_DRIVER_NAME="radeonsi"
            elif lspci -d ::0300 2>/dev/null | grep -qi nvidia; then
              export LIBVA_DRIVER_NAME="nvidia"
            fi

            # Vulkan ICD discovery: tell vulkan-loader where to find GPU drivers
            # NixOS installs ICDs under /run/opengl-driver/share/vulkan/icd.d/
            if [ -d "/run/opengl-driver/share/vulkan/icd.d" ]; then
              export VK_DRIVER_FILES=$(find /run/opengl-driver/share/vulkan/icd.d -name '*.json' 2>/dev/null | tr '\n' ':' | sed 's/:$//')
            fi
          '';
        };
      }
    );
}
