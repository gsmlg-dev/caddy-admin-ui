import "./globals.css";

export const metadata = {
  title: "Caddy Admin",
  description: "Caddy Admin UI",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en" data-theme="dracula">
      <head>
        <link rel="icon" href="/favicon.png" />
      </head>
      <body
        className={`antialiased bg-base-200 text-base-content flex flex-col w-full min-h-screen`}
      >
        {children}
      </body>
    </html>
  );
}
