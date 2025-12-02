/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode() {}
 *     ListNode(int val) { this.val = val; }
 *     ListNode(int val, ListNode next) { this.val = val; this.next = next; }
 * }
 */
class Solution {
    public ListNode partition(ListNode head, int x) {
        //create 2 dummy Nodes for 2 lists 
        ListNode beforeDummy=new ListNode(0);
        ListNode afterDummy=new ListNode(0);
        ListNode before=beforeDummy, after=afterDummy, cur=head;
        // iterate on the main list and update the pointers for seperating them into 2 lists
        while(cur!=null){
            if(cur.val<x){
                before.next=cur;
                before=before.next;
                cur=cur.next;
            } else{
                after.next=cur;
                after=after.next;
                cur=cur.next;
            }
        }
        // ensure that last right node might not be last of the main list
        // which might have a connection to left list node, resulting into circular referrincing
        // so update the last right node.next to null
        after.next=null;
        before.next=afterDummy.next;
        return beforeDummy.next;
    }
}