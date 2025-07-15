"use client";

import { useRef, useEffect } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from '@xterm/addon-fit';

export default () => {
  const terminalRef = useRef();
  useEffect(() => {
    if (!terminalRef.current) return;

    const terminal = new Terminal({
      cursorBlink: true,
      cursorStyle: "block",
      fontSize: 14,
      fontFamily: "Courier New, monospace",
      theme: {
        background: "#000000",
        foreground: "#ffffff",
        cursor: "#ffffff",
        black: "#000000",
        red: "#ff0000",
        green: "#00ff00",
        yellow: "#ffff00",
        blue: "#0000ff",
        magenta: "#ff00ff",
        cyan: "#00ffff",
        white: "#ffffff",
        brightBlack: "#808080",
        brightRed: "#ff8080",
        brightGreen: "#80ff80",
        brightYellow: "#ffff80",
        brightBlue: "#8080ff",
        brightMagenta: "#ff80ff",
        brightCyan: "#80ffff",
        brightWhite: "#ffffff",
      },
      allowTransparency: false,
      convertEol: true,
      scrollback: 10000,
      tabStopWidth: 4,
    });

    // Load the fit addon
    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);

    // Open terminal in the container
    terminalRef.current.innerHTML = "";
    terminal.open(terminalRef.current);

    fitAddon.fit();

    let ws = null;
    let reconnectTimeout = null;

    function updateTerminalSize() {
      const cols = terminal.cols;
      const rows = terminal.rows;

      // Send resize information to server
      if (ws && ws.readyState === WebSocket.OPEN) {
        const resizeData = new Uint8Array(5);
        resizeData[0] = 1; // Resize message marker
        resizeData[1] = (rows >> 8) & 0xff;
        resizeData[2] = rows & 0xff;
        resizeData[3] = (cols >> 8) & 0xff;
        resizeData[4] = cols & 0xff;
        ws.send(resizeData);
      }
    }

    function base64ToUint8(base64) {
      let binary = atob(base64);
      let len = binary.length;
      let bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binary.charCodeAt(i);
      }
      return bytes;
    }
    function uint8ToBase64(uint8Array) {
      let binary = "";
      for (let i = 0; i < uint8Array.length; i++) {
        binary += String.fromCharCode(uint8Array[i]);
      }
      return btoa(binary);
    }
    function encodeUTF8Base64(str) {
      return btoa(String.fromCharCode(...new TextEncoder().encode(str)));
    }
    function decodeUTF8Base64(b64) {
      return new TextDecoder().decode(
        Uint8Array.from(atob(b64), (c) => c.charCodeAt(0)),
      );
    }

    function connect() {
      ws = new WebSocket("/ws/pty");

      ws.onopen = function () {
        // updateStatus(true);
        terminal.write("\r\n\x1b[32mConnected to terminal server\x1b[0m\r\n");
        updateTerminalSize();
      };

      ws.onmessage = function (event) {
        if (event.data instanceof ArrayBuffer) {
          terminal.write(new Uint8Array(event.data));
        } else {
          let data = decodeUTF8Base64(event.data);
          // console.log('data', event.data, data, data.charCodeAt(0), base64ToUint8(event.data));
          if (data.charCodeAt(0) < 0x20) {
            data = base64ToUint8(event.data);
            terminal.write(data);
          } else {
            terminal.write(data);
          }
        }
      };

      ws.onclose = function () {
        // updateStatus(false);
        terminal.write(
          "\r\n\x1b[31mConnection closed. Attempting to reconnect...\x1b[0m\r\n",
        );

        // Attempt to reconnect after 3 seconds
        reconnectTimeout = setTimeout(connect, 3000);
      };

      ws.onerror = function (error) {
        terminal.write("\r\n\x1b[31mWebSocket error: " + error + "\x1b[0m\r\n");
      };
    }

    // Handle terminal input
    terminal.onData(function (data) {
      if (ws && ws.readyState === WebSocket.OPEN) {
        const encoded = encodeUTF8Base64(data);
        ws.send(encoded);
      }
    });

    // Handle terminal resize
    terminal.onResize(function (size) {
      updateTerminalSize();
    });

    // Handle window resize
    window.addEventListener("resize", function () {
      fitAddon.fit();
    });

    // Handle page unload
    window.addEventListener("beforeunload", function () {
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
      if (ws) {
        ws.close();
      }
    });

    // Auto-resize when terminal container changes
    const resizeObserver = new ResizeObserver(function () {
      fitAddon.fit();
    });
    resizeObserver.observe(terminalRef.current);

    // Start connection
    connect();

    // Initial terminal message
    terminal.write("\x1b[36mInitializing terminal emulator...\x1b[0m\r\n");
  }, [terminalRef.current]);
  return (
    <div ref={terminalRef} className="w-full h-full m-4 flex flex-1">
      <div>Loading Ternimal ...</div>
    </div>
  );
};
