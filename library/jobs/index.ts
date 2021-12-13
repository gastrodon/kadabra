import Config from "../types/config";
import { ConfigKind } from "../types/config-kind";
import { JobKind } from "../types/job-kind";

const jobs = new Map([
  [JobKind.QueueLoadStream, require("./queue/load_stream").default],
]);

export default (job: JobKind) => (config: Config) => jobs.get(job)(config);