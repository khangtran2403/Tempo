export interface Workflow {
  id: string;
  user_id: string;
  name: string;
  description: string;
  definition: WorkflowDefinition;
  status: 'draft' | 'active' | 'inactive';
  is_active: boolean;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface WorkflowDefinition {
  trigger: Action;
  actions: Action[];
}

export interface Action {
  id: string;
  type: string; // webhook, http, email, cron
  config: Record<string, any>;
  condition?: Condition;
  transform?: Transform;
}

export interface Condition {
  field: string;
  operator: '==' | '!=' | '>' | '<' | 'contains';
  value: any;
}

export interface Transform {
  input: Record<string, string>;
  template: string;
  output: string;
}
export interface WorkflowVersion{
  ID            :  string          
	WorkflowID    :   string             
	Version       : string               
	Definition    : WorkflowDefinition 
	ChangeSummary :string           
	CreatedBy     : string             
	CreatedAt     : string          
	IsActive      : boolean  
}
