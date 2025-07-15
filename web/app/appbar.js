import Link from "next/link";

const pages = [
  { name: "Config", href: "/" },
  { name: "Upstream", href: "/upstream" },
  { name: "Setup", href: "/setup" },
  { name: "Metrics", href: "/monitor" },
  { name: "SHELL", href: "/shell" },
];

export const AppBar = () => {
  return (
    <div className="navbar bg-primary text-primary-content shadow-sm">
      <div className="navbar-start">
        <div className="dropdown">
          <div tabIndex="0" role="button" className="btn btn-ghost lg:hidden">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M4 6h16M4 12h8m-8 6h16"
              />
            </svg>
          </div>
          <ul
            tabIndex="0"
            className="menu menu-sm dropdown-content bg-primary text-primary-content rounded-box z-1 mt-3 w-52 p-2 shadow"
          >
            {pages.map((p) => (
              <li key={p.name}>
                <Link href={p.href}>{p.name}</Link>
              </li>
            ))}
          </ul>
        </div>
        <Link className="btn btn-ghost text-xl" href="/">
          Caddy Admin
        </Link>
      </div>
      <div className="navbar-center hidden lg:flex">
        <ul className="menu menu-horizontal px-1">
          {pages.map((p) => (
            <li key={p.name}>
              <Link href={p.href}>{p.name}</Link>
            </li>
          ))}
        </ul>
      </div>
      <div className="navbar-end"></div>
    </div>
  );
};
