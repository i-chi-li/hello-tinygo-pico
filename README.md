# hello-tinygo-pico

このプロジェクトは、TinyGo で Raspberry Pi Pico（以降 Pico と呼称する） 用のプログラムを開発するためのサンプルプロジェクトとなる。
開発で利用するツールは、MinGW64 でコンパイルし、cmd コンソール上で実行する前提とする。

Pico は、Debugprobe 用と開発用の 2 つ必要となる。

Pico のプログラム開発には、[Debugprobe](https://github.com/raspberrypi/debugprobe)（旧 Picoprobe） を利用する。
Debugprobe は、[製品版 Raspberry Pi Debug Probe](https://www.raspberrypi.com/products/debug-probe/) と
Pico に書き込んで利用する二通りある。
ここでは、Pico に書き込んで利用する方法を前提とする。
設定方法は、後述する。

# 環境
以下のツール類は、インストール済みとする。

- [Go 言語](https://go.dev/)
- [TinyGo](https://tinygo.org/)
- [MSYS2](https://www.msys2.org/) MinGW64

各ツールのインストールには、[Scoop](https://scoop.sh/) （パッケージ管理ツール）を利用すると便利。
Scoop でインストールしたツールは、`%USERPROFILE%\scoop\apps` ディレクトリ以下に配置される。
以降の手順は、Scoop を利用して上記ツールをインストールした前提で記述する。

IntelliJ は、以下のプラグインをインストールしておく。

- Go
- TinyGo
- Serial Port Monitor

# 事前設定

## プロジェクトをクローンする

```
git clone https://github.com/i-chi-li/hello-tinygo-pico.git
git submodule update --init
```

## Debugprobe の設定

[Debugprobe](https://github.com/raspberrypi/debugprobe) のリリースページから
Debugprobe の `debugprobe_on_pico.uf2` (v2.0.1) ファイルをダウンロードする。
`debugprobe.uf2` ではないので注意。
`debugprobe.uf2` を使用してしまうと、OpenOCD での書き込みで以下のようなエラーとなる。
ネット上の情報では、ブレッドボードを利用したり、配線が長いなどの要因でも発生するようだ。

```
Info : DAP init failed
in procedure 'program'
** OpenOCD init failed **
shutdown command invoked
```

ダウンロードした `debugprobe_on_pico.uf2` ファイルは、
`BOOT SEL` ボタンを押下したまま接続した Pico の USB ドライブに書き込む。

Windows では、Debugprobe を書き込んだ Pico を USB で接続すると、
デバイス マネージャーに、以下のように表示される。

- ユニバーサル シリアル バス デバイス
  - CMSIS-DAP v2 Interface

## Debugprobe と開発用 Pico を接続する

[Pico Getting Started Guide](https://datasheets.raspberrypi.com/pico/getting-started-with-pico.pdf) の
`Appendix A: Using debugprobe`、`Debug with a second Pico` にある接続図の通りに接続する。

Debugprobe 側の Pico に USB を接続する。
開発用 Pico の UART 出力は、Debugprobe 側の USB 接続から COM ポートを介して参照できる。

## シリアルポート接続用ツール

IntelliJ、RLogin や、TeraTerm などのツールを利用する。

IntelliJ を利用する場合は、Serial Port Monitor （JetBrains 製）プラグインをインストールする。
`View` メニュー -> 'Tool Windows' -> `Serial Connections` を選択する。
Serial Connections ツールウィンドウには、`Available Prots` に利用可能な COM ポートが表示される。
参照したい COM ポートを選択する。

- Baud: 115200
- Bits: 8
- Stop bits: 1
- Parity: None
- New line: LF
- Encoding: UTF-8
- Local echo: チェック無し

`Create profile...` をクリックすると、設定した値に名前を付けてプロファイルとして保存できる。

`Connect` ボタンを押下すると、接続開始する。

タブが開き、ログが表示される。
初期表示では、16 進表示なので、左側の `Switch to HEX View` アイコンをクリックして OFF にする。

## picotool をビルドする。
picotool は、Pico に対して、実行ファイルを書き込んだり、
各種情報を取得したりするツール。必須ではない。

picotool のビルドは、MSYS2 MinGW64 上で行う前提とする。
MinGW64 コンソールを起動して、ホームディレクトリに移動する。

```
# picotool で必要なパッケージをインストール（インストール対象の選択肢が表示されるので、そのまま Enter ですべてインストールする）
pacman -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-cmake mingw-w64-x86_64-libusb

export PICO_SDK_PATH=$PWD/pico-sdk
cmake -G"MinGW Makefiles" -DCMAKE_EXE_LINKER_FLAGS="-static -static-libgcc -static-libstdc++" -S picotool -B picotool-build
cmake --build picotool-build -j 12
```

ビルドに成功すると、実行ファイル（`picotool-build/picotool.exe`）が生成される。
`/mingw64/bin/libusb-1.0.dll` が依存 DLL として必要となる。
MinGW64 ターミナル上で実行すれば問題ないが、
cmd 上で実行する場合には、`picotool.exe` ファイルと同じ場所に DLL をコピーするか、
`<MSYS インストールディレクトリ>/mingw64/bin` にパスを通しておく。

## OpenOCD をビルドする

OpenOCD は、設定ファイルなどを含むので、MinGW 上にインストールする。
`<MSYS ディレクトリ>\mingw64\bin` にパスを設定することで、cmd 上からでも呼び出せる。

MSYS2 コンソールを開く。

```
pacman -Syu
pacman -Su
pacman -S mingw-w64-x86_64-toolchain git make libtool pkg-config autoconf automake texinfo mingw-w64-x86_64-libusb
exit
```

MinGW64 コンソールを開く。

```
cd openocd
./bootstrap
./configure --enable-picoprobe --disable-werror
make -j 12
make install
exit
```

## pico-sdk の elf2uf2 をビルド

MinGW64 コンソールを開く。

```
cmake -G"MinGW Makefiles" -S pico-sdk/tools/elf2uf2 -B elf2uf2-build
cmake --build elf2uf2-build -j 12
```

# ビルド

ビルド以降は、cmd.exe ターミナル上で実施する。
以後のコマンド実行には、パスに MinGW64 の bin ディレクトリを追加しておく必要がある。

```
path %USERPROFILE%\scoop\apps\msys2\current\mingw64\bin;%USERPROFILE%\scoop\apps\msys2\current\usr\bin;%PATH%"
```

ビルドが成功すると、bin ディレクトリに `.elf` と `.uf2` ファイルが生成される。

IntelliJ のターミナルの場合は、以下のように設定しておくと、自動的にパスの設定が行われる。

- `File` メニュー -> `Settings...` -> Tools -> Terminal
- `Application Settings`
  - Shell path： `cmd.exe /k "path %USERPROFILE%\scoop\apps\msys2\current\mingw64\bin;%USERPROFILE%\scoop\apps\msys2\current\usr\bin;%PATH%"`

## make を利用する場合

すべてビルドする場合
プロジェクトルートディレクトリで、以下のコマンドを実行する。

```
make
```

blink_led のみビルドする場合

```
cd blink_led
make
```

## 直接ビルドする場合

```
mkdir bin
tinygo build -target pico -o bin/blink_led.elf --serial uart blink_led/main.go
elf2uf2-build\elf2uf2.exe bin/blink_led.elf bin/blink_led.uf2
```

# Pico にプログラムを書き込む

プログラムを書き込む方法には、直接書き込む方法と、Debugprobe 経由で書き込む方法がある。

## Debugprobe 経由の場合

Debugprobe 側に USB ケーブルを接続する。

### make を利用する場合

```
cd blink_led
make load
```

### 直接コマンドを利用する場合

```
openocd.exe -f interface/cmsis-dap.cfg -f target/rp2040.cfg -c "adapter speed 5000" -c "tcl_port disabled" -c "gdb_port disabled" -c "program bin/blink_led.elf verify reset exit"
```

## USB ドライブで書き込む場合

直接書き込む場合は、開発用 Pico を BOOT SEL ボタンを押下したまま、USB ケーブルを接続する。
接続した Pico の USB ドライブに、`.uf2` ファイルを書き込む。
自動的に、USB ドライブが切断され、プログラムが動作する。

## TinyGo で書き込む場合

ビルドと書き込みを同時に行う。
事前に `.elf` や `.uf2` ファイルをビルドする必要は無い。
開発用 Pico に USB ケーブルを接続する。

```
tinygo flash -target=pico blink_led/main.go
```

通常は、Pico が実行モードになっていても、自動的に BOOT SEL モードに切り替わるはずだが、
場合によっては、失敗するので、その場合は、BOOT SEL ボタンを押下したまま接続し直す。

## picotool で書き込む場合

開発用 Pico に USB ケーブルを接続する。

```
picotool-build\picotool reboot -f -u
picotool-build\picotool load bin/blink_led.uf2
picotool-build\picotool reboot
```

通常は、Pico が実行モードになっていても、自動的に BOOT SEL モードに切り替わるはずだが、
場合によっては、失敗するので、その場合は、BOOT SEL ボタンを押下したまま接続し直す。

# デバッグ

MinGW64 の GDB と OpenOCD を利用してデバッグする。
OpenOCD は、Pico と GDB の中継をする。

## make を利用する場合

OpenOCD を起動する。
起動すると、ポート待ち受け状態で待機状態となる。

```
cd blink_led
make openocd
```

別の cmd ターミナルを開く。

```
cd blink_led
make gdb
```

## 直接コマンドを実行する場合

OpenOCD を起動する。
起動すると、ポート待ち受け状態で待機状態となる。

```
openocd.exe -f interface/cmsis-dap.cfg -f target/rp2040.cfg -c "adapter speed 5000"
```

別の cmd ターミナルを開く。

```
cd blink_led
tinygo gdb -target pico
```
