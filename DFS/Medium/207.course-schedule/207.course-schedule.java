class Solution {
    public boolean canFinish(int numCourses, int[][] prerequisites) {
        int []indegree=new int[numCourses];
        List<List<Integer>>graph=new ArrayList<>();
        Queue<Integer>q=new ArrayDeque<>();
        int processed=0;
        for(int node=0;node<numCourses;node++) graph.add(new ArrayList<>());

        for(int []edge:prerequisites){
            int course=edge[0];
            int preq=edge[1];
            indegree[course]++;
            graph.get(preq).add(course);
        }

        for(int node=0;node<indegree.length;node++) 
            if(indegree[node]==0) 
                q.offer(node);

        while(!q.isEmpty()){
            int node=q.poll();
            processed++;
            for(int dep:graph.get(node)){
                indegree[dep]--;
                if(indegree[dep]==0) q.offer(dep);
            }
        }

        return processed==numCourses;
    }
}