import { AppBar } from "../appbar";
import { SaveConfig } from "./save_config";
import { Adapt } from "./adapt";

export default (props) => {
  return (
    <>
      <AppBar />
      <main className="flex flex-col gap-8">
        <Adapt />
        <SaveConfig />
      </main>
    </>
  );
};
