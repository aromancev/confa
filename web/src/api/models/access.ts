import { Store } from "@/platform/store"

export enum Account {
  Guest = 0,
  User = 1,
  Admin = 2,
}

export type Access = {
  id: string
  account: Account
  allowedWrite: boolean
}

export class AccessStore extends Store<Access> {
  login(id: string, acc: Account): void {
    this.reactive.id = id
    this.reactive.account = acc
    this.reactive.allowedWrite = acc === Account.User || acc === Account.Admin
  }
  logout(): void {
    this.reactive.id = ""
    this.reactive.account = Account.Guest
    this.reactive.allowedWrite = false
  }
}

export const accessStore = new AccessStore(defaultState())

function defaultState(): Access {
  return {
    id: "",
    account: Account.Guest,
    allowedWrite: false,
  }
}
