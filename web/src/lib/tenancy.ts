import { get, post, put, remove, sendForm } from "./api";
import type { Usage } from "../components/ui/usage";

export type Organisation = {
  slug: string;
  name: string;
};

export type Project = {
  slug: string;
  name: string;
  domain: string;
  yourRole: string;
  youMay: string[];
  unpublished: number;
  archived?: boolean;
};

export type Member = {
  personId: string;
  email: string;
  displayName: string;
  role: string;
  stillToAccept: boolean;
  /** lastSeen is when they were last here, and absent when they never were. */
  lastSeen?: string;
};

export const roles = ["implementer", "editor", "viewer"] as const;
export type RoleName = (typeof roles)[number];

export function listOrganisations() {
  return get<{ organisations: Organisation[] }>("/api/organisations");
}

export function listProjects(organisation: string) {
  return get<{ projects: Project[]; usage: Usage }>(
    `/api/organisations/${organisation}/projects`,
  );
}

export function listPeople(organisation: string, project: string) {
  return get<{ people: Member[] }>(
    `/api/organisations/${organisation}/projects/${project}/people`,
  );
}

export function invitePerson(organisation: string, project: string, email: string, role: string) {
  return post<Member>(`/api/organisations/${organisation}/projects/${project}/people`, {
    email,
    role,
  });
}

export function changeRole(
  organisation: string,
  project: string,
  personId: string,
  role: string,
) {
  return put<Member>(
    `/api/organisations/${organisation}/projects/${project}/people/${personId}`,
    { role },
  );
}

export function removePerson(organisation: string, project: string, personId: string) {
  return remove<void>(
    `/api/organisations/${organisation}/projects/${project}/people/${personId}`,
  );
}

export type DeployKey = {
  publicKey: string;
  advice: string;
};

export type SiteReport = {
  kind: string;
  remote?: string;
  hosted?: boolean;
  isHugoSite: boolean;
  hasModel: boolean;
  ready: boolean;
  pages: number;
  collections: number;
  problem?: string;
  lookedFor?: string[];
};

export type Connection = {
  good: boolean;
  message: string;
  publicKey?: string;
};

export type Publishing = {
  to: "main-line" | "its-own-line";
  branch?: string;
  form: "source" | "built";
};

export const defaultPublishBranch = "quiqr/publish";

function project(organisation: string, name: string) {
  return `/api/organisations/${organisation}/projects/${name}`;
}

export function readKey(organisation: string, name: string) {
  return get<DeployKey>(`${project(organisation, name)}/key`);
}

export function makeKey(organisation: string, name: string) {
  return post<DeployKey>(`${project(organisation, name)}/key`);
}

export function readSite(organisation: string, name: string) {
  return get<SiteReport>(`${project(organisation, name)}/site`);
}

export function importSite(
  organisation: string,
  name: string,
  body: { kind: string; address: string; remote?: string },
) {
  return post<SiteReport>(`${project(organisation, name)}/import`, body);
}

export function connectSite(organisation: string, name: string, remote: string) {
  return put<SiteReport>(`${project(organisation, name)}/remote`, { remote });
}

export function testConnection(organisation: string, name: string) {
  return post<Connection>(`${project(organisation, name)}/connection`);
}

export function readPublishing(organisation: string, name: string) {
  return get<Publishing>(`${project(organisation, name)}/publishing`);
}

export function setPublishing(organisation: string, name: string, body: Publishing) {
  return put<Publishing>(`${project(organisation, name)}/publishing`, body);
}

export type WordPart = "kept" | "gone" | "added";

export type Word = {
  part: WordPart;
  text: string;
};

export type FieldChange = {
  name: string;
  words?: Word[];
};

export type ChangeKind =
  | "changed-page"
  | "new-page"
  | "removed-page"
  | "new-image"
  | "removed-image"
  | "other-file";

export type Change = {
  kind: ChangeKind;
  section: string;
  page: string;
  by?: string;
  at?: string;
  fields?: FieldChange[];
};

export type Review = {
  count: number;
  lastPublished?: string;
  changes: Change[];
};

export type PublishStep = {
  title: string;
  state: "succeeded" | "failed" | "not-run";
};

export type Publish = {
  id: string;
  state: "running" | "succeeded" | "failed" | "interrupted";
  note?: string;
  changeCount: number;
  startedAt: string;
  finishedAt?: string;
  steps?: PublishStep[];
  trouble?: string;
  detail?: string;
};

export function readReview(organisation: string, name: string) {
  return get<Review>(`${project(organisation, name)}/review`);
}

export function publishNow(organisation: string, name: string, note: string) {
  return post<Publish>(`${project(organisation, name)}/publishes`, { note });
}

export function readPublishes(organisation: string, name: string) {
  return get<{ publishes: Publish[] }>(`${project(organisation, name)}/publishes`);
}

export type Version = {
  version: string;
  title: string;
  by?: string;
  at?: string;
  changes: number;
  live: boolean;
};

export type VersionGroup = {
  period: string;
  entries: Version[];
};

export type VersionTimeline = {
  unpublished: number;
  groups: VersionGroup[];
};

export type GoBackWarning = {
  title: string;
  when?: string;
  setAside: number;
  reassurance: string;
};

export type WentBack = {
  title: string;
  published?: Publish;
};

export function readTimeline(organisation: string, name: string) {
  return get<VersionTimeline>(`${project(organisation, name)}/versions`);
}

export function readVersion(organisation: string, name: string, version: string) {
  return get<Review>(`${project(organisation, name)}/versions/${version}`);
}

export function beforeGoingBack(organisation: string, name: string, version: string) {
  return get<GoBackWarning>(`${project(organisation, name)}/versions/${version}/go-back`);
}

export function goBackTo(
  organisation: string,
  name: string,
  version: string,
  publishRightAway: boolean,
) {
  return post<WentBack>(`${project(organisation, name)}/versions/${version}/go-back`, {
    publishRightAway,
  });
}

export type Option = {
  value: string;
  label: string;
};

export type Control = {
  path: string;
  type: string;
  title: string;
  tip?: string;
  value?: unknown;
  options?: Option[];
  fields?: Control[];
  config?: Record<string, unknown>;
  unknown?: boolean;
};

export type Page = {
  key: string;
  title: string;
  section: string;
  fields: Control[];
  body: string;
  hasBody: boolean;
};

export type Edit = {
  values?: Record<string, unknown>;
  body?: string;
};

export function readPage(organisation: string, name: string, page: string) {
  return get<Page>(`${project(organisation, name)}/pages/${page}`);
}

export function savePage(organisation: string, name: string, page: string, edit: Edit) {
  return put<Page>(`${project(organisation, name)}/pages/${page}`, edit);
}

export type ContentItem = {
  key: string;
  title: string;
  waiting?: boolean;
  by?: string;
  at?: string;
};

export type MenuEntry = {
  key: string;
  kind: "single" | "collection";
  title: string;
  itemName?: string;
  count?: number;
  hasChanges: boolean;
};

export type MenuGroup = { key: string; title: string; items: MenuEntry[] };

export type ContentMenu = {
  groups: MenuGroup[];
  totals: { singles: number; collections: number };
  fallback: boolean;
  warnings?: string[];
};

export type CollectionItem = {
  key: string;
  title: string;
  folder?: string;
  status: "published" | "changed" | "new" | "draft";
  by?: string;
  date?: string;
  author?: string;
  summary?: string;
  thumbnail?: string;
  edited?: string;
  path?: string;
};

export type CollectionSort = "newest" | "oldest" | "title" | "edited" | "model";

export type CollectionItems = {
  key: string;
  title: string;
  itemName: string;
  items: CollectionItem[];
  total: number;
  all: number;
  sort: CollectionSort;
  folders: { name: string; label: string; count: number }[];
  showsPaths: boolean;
  hasImages: boolean;
  around?: { index: number; previous?: string; next?: string };
};

export type ItemsQuery = {
  q?: string;
  folder?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  around?: string;
};

export type ItemAddress = { slug: string; folder?: string; address: string; free: boolean; file?: string };

export type SearchResults = {
  groups: { title: string; items: { key: string; title: string; kind: "single" | "item" }[] }[];
};

function queryString(query: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  for (const [name, value] of Object.entries(query)) {
    if (value !== undefined && value !== "" && value !== 0) params.set(name, String(value));
  }
  const said = params.toString();
  return said ? `?${said}` : "";
}

export function readContentMenu(organisation: string, name: string) {
  return get<ContentMenu>(`${project(organisation, name)}/content-menu`);
}

export function readCollectionItems(organisation: string, name: string, collection: string, query: ItemsQuery = {}) {
  return get<CollectionItems>(`${project(organisation, name)}/collections/${collection}/items${queryString(query)}`);
}

export function readItemAddress(
  organisation: string,
  name: string,
  collection: string,
  asked: { title: string; folder?: string; slug?: string },
) {
  return get<ItemAddress>(`${project(organisation, name)}/collections/${collection}/address${queryString(asked)}`);
}

export function createItem(
  organisation: string,
  name: string,
  collection: string,
  asked: { title: string; folder?: string; slug?: string },
) {
  return post<Page>(`${project(organisation, name)}/collections/${collection}/items`, asked);
}

export function searchContent(organisation: string, name: string, q: string) {
  return get<SearchResults>(`${project(organisation, name)}/search${queryString({ q })}`);
}

export function createPage(organisation: string, name: string, section: string, title: string) {
  return post<Page>(`${project(organisation, name)}/content`, { section, title });
}

export function renamePage(
  organisation: string,
  name: string,
  page: string,
  title: string,
  copy = false,
) {
  return post<Page>(`${project(organisation, name)}/rename/${page}`, { title, copy });
}

export function deletePage(organisation: string, name: string, page: string) {
  return remove<void>(`${project(organisation, name)}/pages/${page}`);
}

export function createProject(
  organisation: string,
  body: { slug: string; name: string; domain?: string },
) {
  return post<Project>(`/api/organisations/${organisation}/projects`, body);
}

export type Image = {
  name: string;
  path: string;
  size: number;
  by?: string;
  at?: string;
  alt?: string;
  usedOn?: string[];
};

export function readImages(organisation: string, name: string) {
  return get<{ images: Image[] }>(`${project(organisation, name)}/images`);
}

export function imageUrl(organisation: string, name: string, image: string, small = false) {
  return `${project(organisation, name)}/images/${image}${small ? "?small=true" : ""}`;
}

export async function addImage(organisation: string, name: string, file: File) {
  const body = new FormData();
  body.append("file", file);
  return sendForm<Image>(`${project(organisation, name)}/images`, body);
}

export function describeImage(organisation: string, name: string, image: string, alt: string) {
  return put<Image>(`${project(organisation, name)}/images/${image}`, { alt });
}

export function removeImage(
  organisation: string,
  name: string,
  image: string,
  evenIfUsed = false,
) {
  return remove<void>(
    `${project(organisation, name)}/images/${image}${evenIfUsed ? "?evenIfUsed=true" : ""}`,
  );
}

export type SyncState = "saved" | "saving" | "offline" | "arrived" | "asking";

export type FieldConflict = {
  path: string;
  name: string;
  mine: string;
  theirs: string;
  words?: Word[];
};

export type PageConflict = {
  page?: string;
  title: string;
  file: string;
  who?: string;
  wholeFile?: boolean;
  fields?: FieldConflict[];
};

export type Standing = {
  state: SyncState;
  said: string;
  unpublished: number;
  arrivedFrom?: string[];
  conflicts?: PageConflict[];
  technical?: { branch: string; remote: string; head: string };
};

export type Answer = {
  file: string;
  fields?: Record<string, "mine" | "theirs">;
  whole?: "mine" | "theirs";
};

export function readStanding(organisation: string, name: string) {
  return get<Standing>(`${project(organisation, name)}/standing`);
}

export function bringChanges(organisation: string, name: string) {
  return post<Standing>(`${project(organisation, name)}/standing/bring`);
}

export function answerChanges(organisation: string, name: string, answers: Answer[]) {
  return post<Standing>(`${project(organisation, name)}/standing/answer`, { answers });
}

export function abandonChanges(organisation: string, name: string) {
  return post<Standing>(`${project(organisation, name)}/standing/abandon`);
}

export function readProject(organisation: string, project: string) {
  return get<Project>(`/api/organisations/${organisation}/projects/${project}`);
}

export function changeProject(
  organisation: string,
  project: string,
  body: { name: string; domain: string },
) {
  return put<Project>(`/api/organisations/${organisation}/projects/${project}`, body);
}

export function removeProject(organisation: string, project: string) {
  return remove<void>(`/api/organisations/${organisation}/projects/${project}`);
}

export type OrganisationPerson = {
  personId: string;
  email: string;
  displayName: string;
  role: string;
  stillToAccept: boolean;
  lastSeen?: string;
  projects: { slug: string; name: string; role: string }[];
};

export type OrganisationArea = {
  slug: string;
  name: string;
  usage: Usage;
  people: OrganisationPerson[];
  projects: Project[];
};

/**
 * readOrganisationArea is everything the organisation's own screens need, in
 * one request: the shell wants the counts whichever screen is open.
 */
export function readOrganisationArea(organisation: string) {
  return get<OrganisationArea>(`/api/organisations/${organisation}/manage`);
}

export function duplicateProject(
  organisation: string,
  project: string,
  body: { slug: string; name: string; withPeople: boolean; withPublishing: boolean },
) {
  return post<Project>(`/api/organisations/${organisation}/projects/${project}/copies`, body);
}

export function moveProject(organisation: string, project: string, to: string) {
  return post<void>(`/api/organisations/${organisation}/projects/${project}/organisation`, { to });
}

export function archiveProject(organisation: string, project: string, archived: boolean) {
  return put<void>(`/api/organisations/${organisation}/projects/${project}/archived`, { archived });
}

/** zipPath is where a site's files are downloaded from. */
export function zipPath(organisation: string, project: string) {
  return `/api/organisations/${organisation}/projects/${project}/zip`;
}

export type WorkspaceStepName = "starting-workspace" | "fetching-site" | "building-preview";

export type WorkspaceStep = {
  step: WorkspaceStepName;
  state: "waiting" | "running" | "done" | "failed";
};

export type Workspace = {
  state: "starting" | "warm" | "sleeping";
  reason?: string;
  steps?: WorkspaceStep[];
  log?: string[];
  address?: string;
  type?: "web" | "pdf" | "other";
  open?: string;
};

export function openWorkspace(organisation: string, name: string) {
  return post<Workspace>(`${project(organisation, name)}/workspace`);
}

export function readWorkspace(organisation: string, name: string) {
  return get<Workspace>(`${project(organisation, name)}/workspace`);
}

export type AssistantChange = {
  page: string;
  title: string;
  before: string;
  after: string;
};

export type AssistantEvent =
  | { type: "ready" }
  | { type: "text"; text: string }
  | { type: "tool"; tool: string; message?: string }
  | { type: "change"; change: AssistantChange }
  | { type: "done" }
  | { type: "error"; message?: string };

export function readAssistant(organisation: string, name: string) {
  return get<{ available: boolean }>(`${project(organisation, name)}/chat`);
}

export type PreviewSelection = {
  text: string;
  before: string;
  after: string;
  path: string;
};

export function askAssistant(organisation: string, name: string, text: string, selection?: PreviewSelection | null) {
  return post<void>(`${project(organisation, name)}/chat`, selection ? { text, selection } : { text });
}

export function assistantEvents(organisation: string, name: string) {
  return `${project(organisation, name)}/chat/events`;
}
