import { AppBar } from "../appbar";
import { Metrics } from "./metrics";

export default (props) => {
  return (
    <>
      <AppBar />
      <div className="card shadow-lg shadow-primary m-4">
        <main className="card-body">
          <h3 className="card-title">Metrics</h3>
          <Metrics />
        </main>
      </div>
    </>
  );
};
