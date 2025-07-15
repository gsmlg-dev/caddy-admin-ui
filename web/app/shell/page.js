import { AppBar } from "../appbar";
import { WebShell } from "./shell";

export default async (props) => {
  return (
    <>
      <AppBar />
      <div className="card shadow-lg shadow-primary m-4 h-full flex-1">
        <main className="card-body">
          <h3 className="card-title">Shell</h3>
          <div className="flex flex-1 h-full w-full">
            <WebShell />
          </div>
        </main>
      </div>
    </>
  );
};
