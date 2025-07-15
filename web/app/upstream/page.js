import { AppBar } from "../appbar";
import { Upstreams } from "./upstream";

export default async (props) => {
  return (
    <>
      <AppBar />
      <div className="card shadow-lg shadow-primary m-4">
        <main className="card-body">
          <h3 className="card-title">Upstream</h3>
          <Upstreams />
        </main>
      </div>
    </>
  );
};
