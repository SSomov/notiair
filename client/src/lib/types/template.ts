export type TemplateDraft = {
	id: string;
	name: string;
	description: string;
	body: string;
	variables: Record<string, string>;
};
