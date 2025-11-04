import type {
	DescMethod,
	DescService,
	Message,
	MessageInitShape,
	MessageShape,
} from "@bufbuild/protobuf";

export type InputOf<
	Desc extends DescService,
	Rpc extends keyof Desc["method"],
> = MessageShape<Desc["method"][Rpc]["input"]>;

export type OutputOf<
	Desc extends DescService,
	Rpc extends keyof Desc["method"],
> = MessageShape<Desc["method"][Rpc]["output"]>;

export type RefOf<
	Desc extends DescService,
	T = MessageInitShape<Desc["method"]["get"]["input"]>,
> = T extends { ref?: any } ? T["ref"] : undefined;

export type QueryDescOf<Desc extends DescService = DescService> = {
	pick: (v: OutputOf<Desc, "get">) => RefOf<Desc>;
	refs: (v: OutputOf<Desc, "get">) => RefOf<Desc>[];
	rpc: {
		[K in keyof Desc["method"]]: {
			desc: Desc["method"][K];
			extract: (
				v: OutputOf<Desc, K>,
			) => OutputOf<Desc, "get"> | OutputOf<Desc, "get">[] | undefined;
		};
	};
};

export type QueryDesc = {
	pick: (v: Message) => { [x: string]: unknown } | undefined;
	refs: (v: Message) => { [x: string]: unknown }[];
	rpc: Record<
		string,
		| undefined
		| {
				desc: DescMethod;
				extract: (v: Message) => Message | Message[] | undefined;
		  }
	>;
};
