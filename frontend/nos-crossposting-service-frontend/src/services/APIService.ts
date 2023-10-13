import axios, {AxiosResponse} from 'axios';
import {CurrentUser} from "@/dto/CurrentUser";
import {Mutation} from '@/store';
import {PublicKeys} from "@/dto/PublicKeys";

export class APIService {

    private readonly axios = axios.create();

    constructor(private store: any) {
    }

    currentUser(): Promise<AxiosResponse<CurrentUser>> {
        const url = `/api/current-user`;
        return this.axios.get<CurrentUser>(url);
    }

    publicKeys(): Promise<AxiosResponse<PublicKeys>> {
        const url = `/api/public-keys`;
        return this.axios.get<PublicKeys>(url);
    }

    refreshCurrentUser(): Promise<CurrentUser> {
        return new Promise((resolve, reject) => {
            this.currentUser()
                .then(
                    response => {
                        this.store.commit(Mutation.SetUser, response.data.user);
                        resolve(response.data);
                    },
                    error => {
                        reject(error);
                    },
                );
        });
    }
}