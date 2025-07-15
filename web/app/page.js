import { AppBar } from "./appbar";
import { Config } from "./config";

export default async function Home({}) {
  return (
    <>
      <AppBar />
      <div className="card shadow-lg shadow-primary m-4">
        <main className="card-body">
          <h3 className="card-title">Config</h3>
          <div className="m-2 p-4 inset-shadow-sm shadow-info">
            <Config />
          </div>
        </main>
      </div>
    </>
  );
}
