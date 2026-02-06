class Solution {
    List<List<Integer>>graph;
    int []state;
    boolean hasCycle;

    public boolean canFinish(int numCourses, int[][] prerequisites) {
        graph=new ArrayList<>();
        state=new int[numCourses]; // state[0]=0 // not visited
        hasCycle=false;
        for(int node=0;node<numCourses;node++) graph.add(new ArrayList<>());
        for(int []edge:prerequisites) graph.get(edge[1]).add(edge[0]);
        for(int node=0;node<numCourses;node++){
            if(state[node]==0) dfs(node);
            if(hasCycle) return false;
        }   
        return true;
    }

    void dfs(int node){
        if(hasCycle) return;
        
        state[node]=1; // visiting 
        for(int nei:graph.get(node)){
            if(state[nei]==1){
                //cycle detected
                hasCycle=true;
                return;
            }
            if(state[nei]==0) dfs(nei);
        }
        state[node]=2;
    }
}